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

package mcp_server_response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

const (
	Name = "mcp-server-response"
)

var (
	_ filter_define.ProxyResponseModifier = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct {
	filter_define.PassThroughResponseModifier
}

var Creator filter_define.ResponseModifierCreator = func(_ string, _ json.RawMessage) filter_define.ProxyResponseModifier {

	return &Filter{}
}

func (f *Filter) OnComplete(resp *http.Response) (out []byte, err error) {
	logrus.Infof("resp: %v", resp.StatusCode)
	return nil, nil
}

func (f *Filter) OnBodyChunk(resp *http.Response, chunk []byte) (out []byte, err error) {
	logger := ctxhelper.MustGetLogger(resp.Request.Context())

	logger.Debugf("method:%v response: %v", resp.Request.Method, string(chunk))

	router, sessionId, err := parseSessionId(string(chunk))
	if err != nil {
		return chunk, nil
	}
	info, ok := ctxhelper.GetmcpMcpInfo(resp.Request.Context())
	if !ok {
		return nil, fmt.Errorf("[Proxy Error] mcp info fail")
	}

	logger.Debugf("Mcp server info: %v", info)

	logger.Infof("message router: %v", router)

	logger.Infof("[Proxy Info] mcp server connect success %v", sessionId)

	return buildMessage(&info, sessionId, router), nil
}

func buildMessage(info *ctxhelper.McpInfo, id string, router string) []byte {
	router = strings.Trim(router, "/")
	buffer := bytes.Buffer{}

	// /proxy/message/{name}/{tag}/{sub-router}?sessionId={sessionId}
	buffer.WriteString("data: /proxy/message/")
	buffer.WriteString(info.Name)
	buffer.WriteString("/")
	buffer.WriteString(info.Version)
	buffer.WriteString("/")
	buffer.WriteString(router)
	buffer.WriteString("?sessionId=")
	buffer.WriteString(id)
	buffer.WriteString("\n")

	return buffer.Bytes()
}

func parseSessionId(message string) (router, sessionId string, err error) {
	re := regexp.MustCompile(`^data:\s*(\/[^\s\?]+)\?sessionId=([0-9a-fA-F-]+)\s*$`)
	matches := re.FindStringSubmatch(message)
	if matches == nil || len(matches) < 3 {
		return "", "", fmt.Errorf("[Proxy Error] failed to parse sessionId")
	}
	return matches[1], matches[2], nil
}
