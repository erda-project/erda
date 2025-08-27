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

package settings

import (
	"errors"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type OpenapiSettings interface {
	GetSessionExpire() time.Duration
}

type config struct {
	SessionExpire time.Duration `file:"session_expire" default:"24h"`
}
type provider struct {
	Cfg *config
}

func (p *provider) GetSessionExpire() time.Duration {
	return p.Cfg.SessionExpire
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	if p.Cfg.SessionExpire < 0 {
		return errors.New("session_expire must be greater than or equal to 0")
	}
	return nil
}

func init() {
	servicehub.Register("openapi-settings", &servicehub.Spec{
		Description: "Openapi global settings",
		Services:    []string{"openapi-settings"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
