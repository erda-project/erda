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

package uc

import (
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Config struct {
	UCClientID     string `default:"dice" file:"UC_CLIENT_ID"`
	UCClientSecret string `default:"secret" file:"UC_CLIENT_SECRET"`
}

type provider struct {
	Cfg *Config
	Log logs.Logger

	baseURL     string
	client      *httpclient.HTTPClient
	ucTokenAuth *UCTokenAuth
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.baseURL = discover.UC()
	p.client = httpclient.New()
	tokenAuth, err := NewUCTokenAuth(p.baseURL, p.Cfg.UCClientID, p.Cfg.UCClientSecret)
	if err != nil {
		return err
	}
	p.ucTokenAuth = tokenAuth
	return nil
}

type Interface common.Interface

func init() {
	servicehub.Register("erda.core.user.uc", &servicehub.Spec{
		Services:   []string{"erda.core.user.uc"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
