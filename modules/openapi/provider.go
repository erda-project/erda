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

package openapi

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/modules/openapi/conf"
)

const serviceName = "openapi"

type provider struct{}

func init() { servicehub.RegisterProvider(serviceName, &provider{}) }

func (p *provider) Service() []string                 { return []string{serviceName} }
func (p *provider) Dependencies() []string            { return []string{} }
func (p *provider) Init(ctx servicehub.Context) error { return nil }
func (p *provider) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}
func (p *provider) Run(ctx context.Context) error {
	logrus.Infof(version.String())
	logrus.Errorf("[alert] openapi instance start")
	conf.Load()
	srv, err := NewServer()
	if err != nil {
		return err
	}
	return srv.ListenAndServe()
}
