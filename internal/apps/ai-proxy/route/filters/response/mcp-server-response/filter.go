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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

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

func (f *Filter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	logger := ctxhelper.MustGetLogger(resp.Request.Context())

	logger.Debugf("method:%v response: %v", resp.Request.Method, string(chunk))

	router, sessionId, err := parseSessionId(string(chunk))
	if err != nil {
		return chunk, nil
	}
	info, ok := ctxhelper.GetMcpInfo(resp.Request.Context())
	if !ok {
		return nil, fmt.Errorf("[Proxy Error] mcp info fail")
	}

	logger.Debugf("Mcp server info: %v", info)

	logger.Infof("message router: %v", router)

	logger.Infof("[Proxy Info] mcp server connect success %v", sessionId)

	return buildMessage(&info, router, chunk), nil
}

func buildMessage(info *ctxhelper.McpInfo, router string, chunk []byte) []byte {
	// prefix：/proxy/message/{name}/{tag}
	prefix := fmt.Sprintf("/proxy/message/%s/%s", info.Name, info.Version)
	prefix = strings.Trim(prefix, "/")
	router = strings.TrimPrefix(router, "/")

	// newRouter：/proxy/message/{name}/{tag}/{router}
	newRouter := "/" + prefix
	if router != "" {
		newRouter = newRouter + "/" + router
	}

	raw := strings.TrimSpace(string(chunk))

	hasDataPrefix := strings.HasPrefix(raw, "data:")
	if hasDataPrefix {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "data:"))
	}

	u, err := url.Parse(raw)
	if err != nil {
		return chunk
	}

	u.Path = newRouter

	out := u.String()
	if hasDataPrefix {
		out = "data: " + out
	}
	out = out + "\n"
	return []byte(out)
}

func parseSessionId(message string) (router, sessionId string, err error) {
	re := regexp.MustCompile(`^(?:data:\s*)?(\/?[^\s\?]+)\?(?:[^ \t\r\n]*&)?(?:sessionId|session_id)=([0-9a-fA-F-]+)(?:&[^\s]*)?\s*$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) < 3 {
		return "", "", fmt.Errorf("[Proxy Error] failed to parse sessionId")
	}

	router, sessionId = matches[1], matches[2]
	if router != "" && router[0] != '/' {
		router = "/" + router
	}
	return router, sessionId, nil
}
