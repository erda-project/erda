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

package iam

import (
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Config struct {
	Host         string `file:"host"`
	ClientID     string `file:"client_id"`
	ClientSecret string `file:"client_secret"`
}

type provider struct {
	Cfg *Config
	Log logs.Logger

	bdl          *bundle.Bundle
	client       *httpclient.HTTPClient
	UserOAuthSvc pb.UserOAuthServiceServer
}

func (p *provider) Init(_ servicehub.Context) error {
	p.client = httpclient.New()
	p.bdl = bundle.New(bundle.WithErdaServer())
	return nil
}

type Interface common.Interface

func init() {
	servicehub.Register("erda.core.user.iam", &servicehub.Spec{
		Services:   []string{"erda.core.user.iam"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
