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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "body-size-limit"
)

var (
	_ reverseproxy.RequestFilter = (*BodySizeLimit)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type BodySizeLimit struct {
	Cfg *Config
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, errors.Wrapf(err, "failed to parse config %s for %s", string(config), Name)
	}
	if len(cfg.Message) == 0 {
		cfg.Message = json.RawMessage(fmt.Sprintf(`{"message": "Request body over length.", "maxSize": %d}`, cfg.MaxSize))
	}
	return &BodySizeLimit{Cfg: &cfg}, nil
}

func (f *BodySizeLimit) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// only check single-modal chat model
	model, _ := ctxhelper.GetModel(ctx)
	if model.Type != modelpb.ModelType_text_generation {
		return reverseproxy.Continue, nil
	}

	var bodyBufferLen int
	if infor.BodyBuffer() != nil {
		bodyBufferLen = infor.BodyBuffer().Len()
	}
	if infor.ContentLength() > f.Cfg.MaxSize || int64(bodyBufferLen) > f.Cfg.MaxSize {
		if ok := json.Valid(f.Cfg.Message); ok {
			w.Header().Set("Content-Type", string(httputil.ApplicationJson))
		}
		_, _ = w.Write(f.Cfg.Message)
		return reverseproxy.Intercept, nil
	}
	return reverseproxy.Continue, nil
}

type Config struct {
	MaxSize int64           `json:"maxSize" yaml:"maxSize"`
	Message json.RawMessage `json:"message" yaml:"message"`
}
