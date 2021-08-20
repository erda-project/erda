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

package cluster_dialer

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/cluster-ops/client"
	"github.com/erda-project/erda/modules/cluster-ops/config"
)

type provider struct {
	Cfg *config.Config
}

func (p *provider) Init(ctx servicehub.Context) error {
	logrus.SetOutput(os.Stdout)

	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	c := client.New(client.WithConfig(p.Cfg))
	return c.Execute()
}

func init() {
	servicehub.Register("cluster-ops", &servicehub.Spec{
		Services:    []string{"cluster-ops"},
		Description: "cluster ops",
		ConfigFunc: func() interface{} {
			return &config.Config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
