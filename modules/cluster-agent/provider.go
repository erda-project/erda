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

package cluster_agent

import (
	"context"

	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/cluster-agent/client"
	"github.com/erda-project/erda/modules/cluster-agent/config"
)

type provider struct {
	Cfg *config.Config // auto inject this field
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		remotedialer.PrintTunnelData = true
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	return client.Start(ctx, p.Cfg)
}

func init() {
	servicehub.Register("cluster-agent", &servicehub.Spec{
		Services:    []string{"cluster-agent"},
		Description: "cluster agent",
		ConfigFunc: func() interface{} {
			return &config.Config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
