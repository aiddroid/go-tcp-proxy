/*
Copyright © 2020 allen <aiddroid@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var whiteIpList WhiteipList
var whiteIpListTicker *time.Ticker
var html *string

type WhiteipList struct {
	Ips []string `json:"ips"`
}

type ProxyCfg struct {
	// 对外暴露的端口, 如443
	FromPort string

	// 对外暴露的端口 8379
	ToPort string

	// 白名单文件路径
	WhiteIpFile string

	// 默认html页面文件路径
	HtmlFile string

	// 新IP认证URI
	AuthUri string

	// 是否dump全部数据
	IsDump bool
}

// 启动代理服务器
func StartServer(config ProxyCfg) {
	address := ":" + config.FromPort
	proxyAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalln("Cannot resolve proxyPort:", config.FromPort)
		return
	}

	proxyListener, err := net.ListenTCP("tcp", proxyAddr)
	if err != nil {
		log.Fatalln("Cannot listen tcp on port:", config.FromPort, err)
		return
	}
	defer proxyListener.Close()

	// 加载IP白名单并设置定时器更新
	loadWhiteIp(config.WhiteIpFile)
	whiteIpListTicker = time.NewTicker(time.Second * 30)
	go func() {
		for {
			select {
				case <-whiteIpListTicker.C: loadWhiteIp(config.WhiteIpFile)
			}
		}
	}()
	
	// 加载HTML
	loadHtml(config.HtmlFile)

	// run
	run(proxyListener, config)
}

// 加载HTML文件
func loadHtml(htmlFile string) {
	if len(htmlFile) > 0 {
		c, err := ioutil.ReadFile(htmlFile)
		if err == nil {
			s := string(c)
			html = &s
			return
		}
	}

	// 构造默认HTML
	defaultHtml := fmt.Sprintf("<html>" +
		"<head><title>Time Page</title></head>" +
		"<body style=\"font-size:12px;\">SERVER TIME: %s</body>" +
		"</html>", time.Now())
	html = &defaultHtml
	return
}

// 加载IP白名单
func loadWhiteIp(whiteIpFile string) {
	bs, err := ioutil.ReadFile(whiteIpFile)
	if err != nil {
		log.Println("Read whitelist file failed:", whiteIpFile)
		return
	}

	json.Unmarshal(bs, &whiteIpList)
	log.Printf("whiteIpList: %s", whiteIpList)
}

// 存储IP到白名单
func storeToWhiteIp(whiteIpFile string, ip string) error {
	whiteIpList.Ips = append(whiteIpList.Ips, ip)
	bs, err := json.MarshalIndent(whiteIpList, "", "    ")
	if err != nil {
		log.Println("Update whitelist failed, ip:", ip)
		return err
	}

	if e := ioutil.WriteFile(whiteIpFile, bs, 0775); e != nil {
		return e
	}

	log.Println("Update whitelist succeed, ip:", ip)
	return nil
}

// run
func run(proxyListener *net.TCPListener, config ProxyCfg) {
	for {
		// 持续监听新请求
		proxyConn, err := proxyListener.AcceptTCP()
		if err != nil {
			log.Fatalln("Cannot accept tcp connection via port:", config.FromPort)
			return
		}

		// 保持连接
		proxyConn.SetKeepAlive(true)
		proxyConn.SetKeepAlivePeriod(time.Minute)

		// 获取客户端IP
		clientAddr := proxyConn.RemoteAddr()
		// 从clientAddr中解析IP
		clientIp := parseClientIp(clientAddr)
		log.Println("Client connected from ip:", clientIp)

		// 对不在白名单中的ip，进行特殊处理
		if len(whiteIpList.Ips) > 0 && !funk.ContainsString(whiteIpList.Ips, clientIp){
			// 只读取HTTP请求前100个字节
			buffer := make([]byte, 100)
			n, err := proxyConn.Read(buffer)
			if err == nil {
				// 如果是AUTH请求，则把ip加到白名单
				if strings.Index(string(buffer[:n]), "GET " + config.AuthUri) == 0 {
					if e := storeToWhiteIp(config.WhiteIpFile, clientIp); e == nil {
						proxyConn.Write([]byte(httpResp("SUCCESS")))
					} else {
						proxyConn.Write([]byte(httpResp("FAILED")))
					}
				} else if strings.Index(string(buffer[:n]), "GET /") == 0 {
					// 如果是其他GET请求，则直接返回html
					resp := httpResp(*html)

					// proxyConn.SetNoDelay(true)
					proxyConn.Write([]byte(resp))

					log.Printf("Response HTML to ip:%s", clientIp)
					log.Printf("Filtered ip: %s", clientIp)
				}
			}
			proxyConn.Close()
			continue
		}

		// target
		address := ":" + config.ToPort
		targetAddr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("Cannot resolve targetPort:", config.ToPort)
			return
		}

		targetConn, err := net.DialTCP("tcp", nil, targetAddr)
		if err != nil {
			log.Println("Cannot connect to target port:", config.ToPort)
			continue
		}

		targetConn.SetKeepAlive(true)
		targetConn.SetKeepAlivePeriod(time.Hour)


		// log.Println("goroutine Id:", GoId())

		// 读取客户端数据发送给目标
		go doProxy(proxyConn, targetConn, config.IsDump, true)

		// 读取目标数据响应，返回给客户端
		go doProxy(targetConn, proxyConn, config.IsDump, false)
	}
}

// 从clientAddr中解析出IP
func parseClientIp(clientAddr net.Addr) string {
	// loopback IP  [::1]:60940
	if strings.Contains(clientAddr.String(), "::") {
		return clientAddr.String()[1:3]
	}

	return strings.Split(clientAddr.String(), ":")[0]
}

// 构建http响应
func httpResp(body string) string {
	http := "HTTP/1.1 200 OK\r\n"
	http += "Content-Type: text/html\r\n"
	http += "Content-Length: %d\r\n\r\n"
	http += "%s"
	return fmt.Sprintf(http, len(body), body)
}

// 在sockets之间做代理转发
func doProxy(readConn *net.TCPConn, writeConn *net.TCPConn, isDump bool, isProxy bool) {
	defer readConn.Close()
	defer writeConn.Close()

	// 每次读取 500KB
	buffer := make([]byte, 1024 * 500)
	for {
		n, err := readConn.Read(buffer)
		if err != nil {
			break
		}

		// log.Printf("Read %d bytes from %s", n, Conditional(isProxy, "client", "upstream"))
		// log.Printf("read from %p: %s", readConn, Conditional(isProxy, "client", "upstream"))
		// log.Printf("write to %p: %s", writeConn, Conditional(isProxy, "client", "upstream"))

		if isDump {
			log.Printf("dump raw data:%s", buffer[:n])
		}

		n, err = writeConn.Write(buffer[:n])
		if err != nil {
			break
		}

		// log.Printf("Write %d bytes to %s", n, Conditional(isProxy, "upstream", "client"))
	}

}

// conditional check
func Conditional(condition bool, value1 interface{}, value2 interface{}) interface{} {
	if condition {
		return value1
	}

	return value2
}

func GoId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
