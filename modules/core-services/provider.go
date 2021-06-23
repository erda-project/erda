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

package core_services

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type provider struct{}

func (p *provider) Run(ctx context.Context) error { return Initialize() }

func init() {
	servicehub.Register("core-services", &servicehub.Spec{
		Services: []string{"core-services"},
		Creator:  func() servicehub.Provider { return &provider{} },
	})
}
