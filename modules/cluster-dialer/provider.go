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

package cluster_dialer

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/modules/cluster-dialer/config"
	"github.com/erda-project/erda/modules/cluster-dialer/server"
)

type provider struct {
	Cfg        *config.Config             // auto inject this field
	Credential tokenpb.TokenServiceServer `autowired:"erda.core.token.TokenService" optional:"true"`
	Etcd       *clientv3.Client           `autowired:"etcd"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		remotedialer.PrintTunnelData = true
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	return server.Start(ctx, p.Credential, p.Cfg, p.Etcd)
}

func init() {
	servicehub.Register("cluster-dialer", &servicehub.Spec{
		Services:    []string{"cluster-dialer"},
		Description: "cluster dialer",
		ConfigFunc:  func() interface{} { return &config.Config{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
