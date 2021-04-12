// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package http

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

//func TestHTTP(t *testing.T) {
//	kill, err := RunAHTTPServer(12345, "content")
//	assert.Nil(t, err)
//	defer kill()
//	director := func(r *http.Request) {
//		r.URL.Scheme = "http"
//		r.URL.Host = "127.0.0.1:12345"
//		r.Header.Set("Origin", "http://127.0.0.1:12345")
//	}
//	p := NewReverseProxy(director, nil)
//	go http.ListenAndServe("127.0.0.1:6666", p)
//	time.Sleep(1000 * time.Millisecond)
//	r, err := http.Get("http://127.0.0.1:6666")
//	assert.Nil(t, err)
//	body, err := ioutil.ReadAll(r.Body)
//	assert.Nil(t, err)
//	assert.Equal(t, "content", string(body))
//
//}

// return (kill_function, error)
func RunAHTTPServer(port int, content string) (func() error, error) {
	content_ := strings.Replace(content, "\"", "\\\"", -1)
	socat := fmt.Sprintf("socat tcp-listen:%d,reuseaddr,fork \"exec:printf \\'HTTP/1.0 200 OK\\r\\nContent-Type\\:application/json\\r\\n\\r\\n%s\\'\"", port, content_)
	cmd := exec.Command("sh", "-c", socat)
	err := cmd.Start()
	time.Sleep(500 * time.Millisecond)
	return cmd.Process.Kill, err
}
