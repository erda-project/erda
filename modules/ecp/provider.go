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

package ecp

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/ecp/conf"
)

type provider struct {
	Cfg *conf.Conf
}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	return p.initialize()
}

func init() {
	servicehub.Register("ecp", &servicehub.Spec{
		Services:    []string{"ecp"},
		Description: "Core components of edge computing platform.",
		ConfigFunc:  func() interface{} { return &conf.Conf{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
