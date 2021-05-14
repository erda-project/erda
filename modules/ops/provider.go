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

// Package ops Core components of multi-cloud management platform
package ops

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type provider struct{}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	logrus.Info("ops provider is running...")
	return initialize()
}

func init() {
	servicehub.Register("ops", &servicehub.Spec{
		Services:    []string{"ops"},
		Description: "Core components of multi-cloud management platform.",
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
