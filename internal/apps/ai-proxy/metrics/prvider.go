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

package metrics

import (
	"reflect"

	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

var (
	Name         = "erda.app.ai-proxy.metrics"
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.ai-proxy.metrics.Collectors"},
		Summary:     "ai-proxy prometheus collectors",
		Description: "ai-proxy prometheus collectors",
		Types:       []reflect.Type{providerType},
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(Name, &spec)
}

type provider struct {
	L logs.Logger
	D *gorm.DB `autowired:"mysql-gorm.v2-client"`
}

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return SingletonCollector(p.D, p.L)
}
