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
	"io"
	"net/http"

	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "context-file"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Context struct {
	Config *Config
}

type Config struct {
}

func New(configJSON json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}
	return &Context{Config: &cfg}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	originBody := infor.BodyBuffer(true)
	defer infor.SetBody(io.NopCloser(originBody), int64(originBody.Len()))

	// parse multiform/data
	_, fileHeader, err := infor.Request().FormFile("file")
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse file field, err: %v", err)
	}
	purpose := infor.Request().FormValue("purpose")
	if purpose == "" {
		return reverseproxy.Intercept, fmt.Errorf("purpose is required")
	}
	// use purpose as prompt
	prompt := fmt.Sprintf("filename: %s, purpose: %s", fileHeader.Filename, purpose)
	ctxhelper.PutUserPrompt(ctx, prompt)

	return reverseproxy.Continue, nil
}
