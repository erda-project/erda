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

package body_size_limit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

const (
	Name = "body-size-limit"
)

var (
	_ filter_define.ProxyRequestRewriter = (*BodySizeLimit)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type BodySizeLimit struct {
	MaxSize int64           `json:"maxSize" yaml:"maxSize"`
	Message json.RawMessage `json:"message" yaml:"message"`
}

var Creator filter_define.RequestRewriterCreator = func(_ string, config json.RawMessage) filter_define.ProxyRequestRewriter {
	var f BodySizeLimit
	if err := json.Unmarshal(config, &f); err != nil {
		panic(fmt.Sprintf("failed to parse config %s for %s", string(config), Name))
	}
	if len(f.Message) == 0 {
		f.Message = json.RawMessage(fmt.Sprintf(`{"message": "Request body over length.", "maxSize": %d}`, f.MaxSize))
	}
	return &f
}

func (f *BodySizeLimit) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// only check single-modal chat model
	model, _ := ctxhelper.GetModel(pr.Out.Context())
	if model.Type != modelpb.ModelType_text_generation {
		return nil
	}

	var bodyBufferLen int64
	if pr.In.ContentLength > 0 {
		bodyBufferLen = pr.In.ContentLength
	}
	if bodyBufferLen > f.MaxSize {
		return http_error.NewHTTPError(http.StatusBadRequest, string(f.Message))
	}
	return nil
}
