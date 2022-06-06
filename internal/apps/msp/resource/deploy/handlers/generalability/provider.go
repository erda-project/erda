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

package generalability

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	dicehubpb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/modules/apps/msp/resource/deploy/handlers"
)

type config struct {
}

// +provider
type provider struct {
	*handlers.DefaultDeployHandler
	Cfg               *config
	Log               logs.Logger
	DB                *gorm.DB                       `autowired:"mysql-client"`
	DicehubReleaseSvc dicehubpb.ReleaseServiceServer `autowired:"erda.core.dicehub.release.ReleaseService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultDeployHandler = handlers.NewDefaultHandler(p.DB, p.Log)
	return nil
}

func init() {
	servicehub.Register("erda.msp.resource.deploy.handlers.generalability", &servicehub.Spec{
		Services: []string{
			"erda.msp.resource.deploy.handlers.generalability",
		},
		Description: "erda.msp.resource.deploy.handlers.generalability",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
