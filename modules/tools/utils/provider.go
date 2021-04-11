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

package utils

import (
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	kubernetesp "github.com/erda-project/erda-infra/providers/kubernetes"
)

type define struct{}

func (d *define) Service() []string      { return []string{"monitor-tools"} }
func (d *define) Dependencies() []string { return []string{"http-server", "kubernetes"} }
func (d *define) Summary() string        { return "monitor tools" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

type provider struct {
	C   *config
	L   logs.Logger
	k8s *kubernetes.Clientset
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.k8s = ctx.Service("kubernetes").(kubernetesp.Interface).Client()
	return nil
}

// Start .
func (p *provider) Start() error {
	return p.showK8sNodeResources()
}

func (p *provider) Close() error {
	return nil
}

func init() {
	servicehub.RegisterProvider("monitor-tools", &define{})
}
