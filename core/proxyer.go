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
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

var whiteIpList WhiteipList
var whiteIpListTicker *time.Ticker
var filteredHtml string

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

	// 定时重新加载IP白名单
	loadWhiteIp(config.WhiteIpFile)
	//whiteIpListTicker = time.NewTicker(time.Minute)
	//go func() {
	//	for {
	//		select {
	//			case <- whiteIpListTicker.C:
	//				loadWhiteIp(config.WhiteIpFile)
	//		}
	//	}
	//}()

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
			filteredHtml = string(c)
			return
		}
	}

	// 构造默认HTML
	defaultHtml := fmt.Sprintf("<html>"+
		"<head><title>Time Page</title></head>"+
		"<body style=\"font-size:12px;\">SERVER TIME: %s</body>"+
		"</html>", time.Now())
	filteredHtml = defaultHtml
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
		clientConn, err := proxyListener.AcceptTCP()
		if err != nil {
			log.Fatalln("Cannot accept tcp connection via port:", config.FromPort, err)
			return
		}

		go handle(clientConn, config)
	}
}

// 处理请求
func handle(clientConn *net.TCPConn, config ProxyCfg) {
	defer clientConn.Close()

	// 获取客户端IP
	clientAddr := clientConn.RemoteAddr()
	// 从clientAddr中解析IP
	clientIp := ParseClientIp(clientAddr)
	log.Println("Client connected from ip:", clientIp)

	// 对不在白名单中的ip，进行特殊处理
	if len(whiteIpList.Ips) > 0 && !funk.ContainsString(whiteIpList.Ips, clientIp) {
		handleNewIpClient(clientConn, clientIp, config)
		return
	}

	// remote
	address := ":" + config.ToPort
	remoteAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalln("Cannot resolve remotePort:", config.ToPort)
		return
	}

	remoteConn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		log.Println("Cannot connect to remote port:", config.ToPort)
		return
	}
	defer remoteConn.Close()

	//clientConn.SetKeepAlive(true)
	//remoteConn.SetKeepAlive(true)

	// log.Println("goroutine Id:", GoId())
	sync := make(chan bool, 2)
	// 读取客户端数据发送给目标
	transfer(clientConn, remoteConn, sync, config, true)
	// 读取目标数据响应，返回给客户端
	transfer(remoteConn, clientConn, sync, config, false)
	log.Println("synced")
}

// 转发数据
func transfer(srcConn *net.TCPConn, dstConn *net.TCPConn, sync chan bool, config ProxyCfg, isClient bool) {
	log.Println(Conditional(isClient, "req: client => proxy => target", "resp: client <= proxy <= target"))

	srcConn.SetReadDeadline(time.Now().Add(time.Second * 5))
	// read 1024KB=1MB
	size := 1024 * 1024
	for {
		bs := make([]byte, size)
		nr, er := srcConn.Read(bs)
		if er != nil {
			sync <- false
			log.Println("read err:", er)
			return
		}

		if config.IsDump {
			log.Println("raw data:", string(bs))
		}

		_, ew := dstConn.Write(bs)
		if ew != nil {
			sync <- false
			log.Println("write err:", ew)
			return
		}

		if nr < size {
			sync <- true
			return
		}
	}
}

// 处理新IP客户端
func handleNewIpClient(proxyConn *net.TCPConn, clientIp string, config ProxyCfg) {
	defer proxyConn.Close()

	proxyConn.SetReadDeadline(time.Now().Add(time.Second * 3))
	// 只读取HTTP请求前100个字节
	buffer := make([]byte, 100)
	n, err := proxyConn.Read(buffer)
	if err != nil {
		return
	}

	// 如果是AUTH请求，则把ip加到白名单
	if strings.Index(string(buffer[:n]), "GET "+config.AuthUri) == 0 {
		if e := storeToWhiteIp(config.WhiteIpFile, clientIp); e == nil {
			// 重新加载白名单
			loadWhiteIp(config.WhiteIpFile)
			proxyConn.Write([]byte(HttpResp("SUCCESS " + clientIp)))
			return
		}
		proxyConn.Write([]byte(HttpResp("FAILED " + clientIp)))
	} else if strings.Index(string(buffer[:n]), "GET /") == 0 {
		// 如果是其他GET请求，则直接返回html
		resp := HttpResp(filteredHtml)

		// proxyConn.SetNoDelay(true)
		proxyConn.Write([]byte(resp))

		log.Printf("Response html to unknow ip:%s", clientIp)
		return
	}
}
