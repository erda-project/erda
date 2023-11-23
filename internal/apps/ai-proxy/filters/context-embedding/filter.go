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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "context-embedding"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Context struct {
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Context{}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var req openai.EmbeddingRequest
	if err := json.NewDecoder(infor.BodyBuffer()).Decode(&req); err != nil {
		return reverseproxy.Continue, fmt.Errorf("failed to decode embedding request, err: %v", err)
	}
	switch req.Input.(type) {
	case string:
		ctxhelper.PutUserPrompt(ctx, req.Input.(string))
	case []interface{}:
		var ss []string
		for _, item := range req.Input.([]interface{}) {
			ss = append(ss, strutil.String(item))
		}
		ctxhelper.PutUserPrompt(ctx, strutil.Join(ss, "\n"))
	}
	return reverseproxy.Continue, nil
}
