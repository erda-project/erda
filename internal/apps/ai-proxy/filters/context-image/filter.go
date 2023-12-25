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
	"strings"

	"github.com/sashabaranov/go-openai"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "context-image"
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
	SupportedResponseFormats []string `json:"supportedResponseFormats" yaml:"supportedResponseFormats"`
}

func New(configJSON json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}
	return &Context{Config: &cfg}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// check request
	var req openai.ImageRequest
	if err := json.NewDecoder(infor.BodyBuffer()).Decode(&req); err != nil {
		return reverseproxy.Intercept, err
	}
	// prompt
	if strings.TrimSpace(req.Prompt) == "" {
		return reverseproxy.Intercept, fmt.Errorf("prompt is empty")
	}
	ctxhelper.PutUserPrompt(ctx, req.Prompt)
	// response format
	if !strutil.Exist(f.Config.SupportedResponseFormats, req.ResponseFormat) {
		return reverseproxy.Intercept, fmt.Errorf("unsupported response format: %s (supported: %v)", req.ResponseFormat, f.Config.SupportedResponseFormats)
	}

	// put image info
	imageInfo := ctxhelper.ImageInfo{
		ImageQuality: req.Quality,
		ImageSize:    req.Size,
		ImageStyle:   req.Style,
	}
	ctxhelper.PutImageInfo(ctx, imageInfo)

	return reverseproxy.Continue, nil
}
