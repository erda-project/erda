// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apigateway

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
)

type config struct {
	MainClusterInfo struct {
		Name       string `file:"name"`
		RootDomain string `file:"root_domain"`
		Protocol   string `file:"protocol"`
		HttpPort   string `file:"http_port"`
		HttpsPort  string `file:"https_port"`
	} `file:"main_cluster_info"`
}

// +provider
type provider struct {
	*handlers.DefaultDeployHandler
	Cfg *config
	Log logs.Logger
	DB  *gorm.DB `autowired:"mysql-client"`
}

// Init this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultDeployHandler = handlers.NewDefaultHandler(p.DB, p.Log)
	return nil
}

func init() {
	servicehub.Register("erda.msp.resource.deploy.handlers.apigateway", &servicehub.Spec{
		Services: []string{
			"erda.msp.resource.deploy.handlers.apigateway",
		},
		Description: "erda.msp.resource.deploy.handlers.apigateway",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
