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

package cluster_manager

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/cluster-manager/conf"
)

type provider struct {
	Cfg *conf.Conf
}

func (p *provider) Init(ctx servicehub.Context) error {
	return initialize(p.Cfg)
}

func (p *provider) Run(ctx context.Context) error {
	return nil
}

func init() {
	servicehub.Register("cluster-manager", &servicehub.Spec{
		Services:    []string{"cluster-manager"},
		Description: "cluster manager",
		ConfigFunc: func() interface{} {
			return &conf.Conf{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
