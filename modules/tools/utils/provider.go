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
