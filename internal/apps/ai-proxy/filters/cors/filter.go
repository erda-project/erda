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

package erda_auth

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"gopkg.in/yaml.v3"
	"net/http"
)

const (
	Name = "cors"
)

var (
	_ reverseproxy.ResponseFilter = (*Cors)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Cors struct {
	*reverseproxy.DefaultResponseFilter

	Config *Config
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &Cors{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
		Config:                &cfg,
	}, nil
}

func (f *Cors) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// todo: get referer
	return reverseproxy.Continue, nil
}

func (f *Cors) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		l.Errorf("failed to f.DefaultResponseFilter.OnResponseEOF, err: %v", err)
		return err
	}

	// todo: check refer in r.Config.AccessControlAllowOrigin

	infor.Header().Set("Access-Control-Allow-Origin", "*")
	return nil
}

type Config struct {
	AccessControlAllowOrigin []string `json:"accessControlAllowOrigin" yaml:"accessControlAllowOrigin"`
}
