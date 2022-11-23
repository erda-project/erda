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

package over_permission

import (
	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/org"
)

type config struct {
	Weight          int64    `file:"weight" default:"50"`
	DefaultMatchOrg []string `file:"default_match_org" default:""`
}

// +provider
type provider struct {
	Cfg   *config
	Log   logs.Logger
	Redis *redis.Client `autowired:"redis-client"`
	Org   org.Interface
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	return nil
}

var _ openapiauth.AutherLister = (*provider)(nil)

func (p *provider) Authers() []openapiauth.Auther {
	return []openapiauth.Auther{newOverPermissionOrg(p)}
}

func init() {
	servicehub.Register("openapi-over-permission", &servicehub.Spec{
		Services:   []string{"openapi-over-permission"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
