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

package cluster_manager

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/jinzhu/gorm"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core/cluster-manager/cluster"
	"github.com/erda-project/erda/modules/core/cluster-manager/cluster/db"
	"github.com/erda-project/erda/modules/core/cluster-manager/conf"
	"github.com/erda-project/erda/modules/core/cluster-manager/dialer/server"
)

type provider struct {
	Cfg        *conf.Conf
	Bdl        *bundle.Bundle
	DB         *gorm.DB                   `autowired:"mysql-client"`
	Router     httpserver.Router          `autowired:"http-router"`
	Credential tokenpb.TokenServiceServer `autowired:"erda.core.token.TokenService" optional:"true"`
	Etcd       *clientv3.Client           `autowired:"etcd"`
	Cluster    clusterpb.ClusterServiceServer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Bdl = bundle.New(bundle.WithCoreServices())
	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		remotedialer.PrintTunnelData = true
	}
	p.Cluster = cluster.NewClusterService(cluster.WithDB(&db.ClusterDB{DB: p.DB}), cluster.WithBundle(p.Bdl))
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.Router.Any("/**", server.NewDialerRouter(ctx, p.Cluster, p.Credential, p.Cfg, p.Etcd).ServeHTTP)
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
