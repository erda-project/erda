// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/modules/eventbox/api"
	"github.com/erda-project/erda/modules/eventbox/conf"
	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/subscriber/fake"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/jsonstore"
)

var originListenAddr string

func init() {
	originListenAddr = conf.ListenAddr()
}

var sender = "eventbox-self"
var notifier, _ = api.New(sender, nil)

func GenContent(content string) string {
	return content + fmt.Sprintf("%d", time.Now().UnixNano())
}

func InputEtcd(content string) error {
	return notifier.Send(content, api.WithDest(map[string]interface{}{
		"FAKE": ""}))
}

func InputHTTP(content string) (*httpclient.Response, *bytes.Buffer, error) {
	return InputHTTPRaw(content, originListenAddr[1:], map[types.LabelKey]interface{}{"FAKE": ""})
}

func InputHTTPRaw(content string, port string, dest map[types.LabelKey]interface{}) (*httpclient.Response, *bytes.Buffer, error) {
	var buf bytes.Buffer
	resp, err := httpclient.New().Post("127.0.0.1:"+port).Path("/api/dice/eventbox/message/create").Header("Accept", "application/json").JSONBody(types.Message{
		Sender:  sender,
		Labels:  dest,
		Content: content,
		Time:    time.Now().UnixNano(),
	}).Do().Body(&buf)
	return resp, &buf, err
}

var (
	fakefile     *os.File
	fakefileLock sync.Once
)

func FakeFileContains(content string) bool {
	s, err := ioutil.ReadFile(fake.FakeTestFilePath)
	if err != nil {
		panic(err)
	}
	all := string(s)
	return strings.Contains(all, content)
}

func IsCleanEtcd() bool {
	js, err := jsonstore.New()
	if err != nil {
		panic(err)
	}
	r, err := js.ListKeys(context.TODO(), constant.MessageDir)
	if err != nil {
		panic(err)
	}
	return len(r) == 0
}

// return (kill_function, error)
func RunAHTTPServer(port int, content string) (func() error, error) {
	content_ := strings.Replace(content, "\"", "\\\"", -1)
	socat := fmt.Sprintf("socat tcp-listen:%d,reuseaddr,fork \"exec:printf \\'HTTP/1.0 200 OK\\r\\nContent-Type\\:application/json\\r\\n\\r\\n%s\\'\"", port, content_)
	cmd := exec.Command("sh", "-c", socat)
	err := cmd.Start()
	time.Sleep(500 * time.Millisecond)
	return cmd.Process.Kill, err
}
