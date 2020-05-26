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
	"net"
	"runtime"
	"strconv"
	"strings"
)

// 从clientAddr中解析出IP
func ParseClientIp(clientAddr net.Addr) string {
	// loopback IP  [::1]:60940
	if strings.Contains(clientAddr.String(), "::") {
		return clientAddr.String()[1:3]
	}

	return strings.Split(clientAddr.String(), ":")[0]
}

// 构建http响应
func HttpResp(body string) string {
	http := "HTTP/1.1 200 OK\r\n"
	http += "Content-Type: text/filteredHtml\r\n"
	http += "Content-Length: %d\r\n\r\n"
	http += "%s"
	return fmt.Sprintf(http, len(body), body)
}

// conditional check
func Conditional(condition bool, value1 interface{}, value2 interface{}) interface{} {
	if condition {
		return value1
	}

	return value2
}

// 获取go routineId
func GoId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
