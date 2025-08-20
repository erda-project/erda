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

package context

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

const (
	Name = "context-file"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// upload file
	if ctxhelper.MustGetPathMatcher(pr.In.Context()).Pattern == common.RequestPathPrefixV1Files && pr.In.Method == http.MethodPost {
		if err := onUploadFile(pr); err != nil {
			return err
		}
	}

	return nil
}

func onUploadFile(pr *httputil.ProxyRequest) error {
	// parse multiform/data
	_, fileHeader, err := pr.In.FormFile("file")
	if err != nil {
		return fmt.Errorf("failed to parse file field, err: %v", err)
	}
	purpose := pr.In.FormValue("purpose")
	if purpose == "" {
		return fmt.Errorf("purpose is required")
	}
	// use purpose as prompt
	prompt := fmt.Sprintf("filename: %s, purpose: %s", fileHeader.Filename, purpose)

	if sink, ok := ctxhelper.GetAuditSink(pr.In.Context()); ok {
		sink.Note("prompt", prompt)
	}

	return nil
}
