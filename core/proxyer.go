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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Start proxy server
func StartServer(proxyPort string, targetPort string, whiteIpFile string) {
	address := ":" + proxyPort
	proxyAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalln("Cannot resolve proxyPort:", proxyPort)
		return
	}

	proxyListener, err := net.ListenTCP("tcp", proxyAddr)
	if err != nil {
		log.Fatalln("Cannot listen tcp on port:", proxyPort, err)
		return
	}
	defer proxyListener.Close()

	whiteIpList, err := ioutil.ReadFile(whiteIpFile)
	if err != nil {
		log.Println("Read whitelist file failed:", whiteIpFile)
	}
	log.Printf("whiteIpList: %s", whiteIpList)

	run(proxyListener, proxyPort, targetPort, whiteIpList)
}

func run(proxyListener *net.TCPListener, proxyPort string, targetPort string, whiteIpList []byte) {
	for {
		proxyConn, err := proxyListener.AcceptTCP()
		if err != nil {
			log.Fatalln("Cannot accept tcp connection via port:", proxyPort)
			return
		}

		proxyConn.SetKeepAlive(true)
		proxyConn.SetKeepAlivePeriod(time.Minute)

		clientAddr := proxyConn.RemoteAddr()
		var clientIp = ""
		if strings.Contains(clientAddr.String(), "::") {
			clientIp = clientAddr.String()[1:3]
		} else {
			clientIp = strings.Split(clientAddr.String(), ":")[0]
		}

		log.Println("Client connected from ", clientAddr, " ip:", clientIp)
		if len(whiteIpList) > 0 && !strings.Contains(string(whiteIpList), clientIp){
			html := fmt.Sprintf("<html><body>SERVER TIME: %s</body></html>", time.Now())
			resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n%s", len(html), html)
			// proxyConn.SetNoDelay(true)
			proxyConn.Write([]byte(resp))

			log.Printf("Filtered ip %s with HTTP", clientIp)
			continue
		}

		// target
		address := ":" + targetPort
		targetAddr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatalln("Cannot resolve targetPort:", targetPort)
			return
		}

		targetConn, err := net.DialTCP("tcp", nil, targetAddr)
		if err != nil {
			log.Println("Cannot connect to target port:", targetPort)
			continue
		}

		targetConn.SetKeepAlive(true)
		targetConn.SetKeepAlivePeriod(time.Minute)


		// log.Println("goroutine Id:", GoId())

		// read client data and send to target
		go doProxy(proxyConn, targetConn, true)
		// read target response data and reply to client
		go doProxy(targetConn, proxyConn, false)
	}
}

// do proxy between sockets
func doProxy(readConn *net.TCPConn, writeConn *net.TCPConn, isProxy bool) {
	defer readConn.Close()
	defer writeConn.Close()

	// read 100KB
	buffer := make([]byte, 1024 * 100)
	for {
		n, err := readConn.Read(buffer)
		if err != nil {
			log.Printf("Cannot Read data, error:%s", err)
			break
		}

		log.Printf("Read %d bytes from %s", n, Conditional(isProxy, "client", "upstream"))
		// log.Printf("data:%s", buffer[:n])

		n, err = writeConn.Write(buffer[:n])
		if err != nil {
			log.Printf("Cannot write data, error:%s", err)
			break
		}

		log.Printf("Write %d bytes to %s", n, Conditional(isProxy, "upstream", "client"))
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
