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

package queuemanage

import (
	"context"
	"fmt"
	"reflect"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanage/manager"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanage/types"
	"github.com/erda-project/erda/pkg/jsonstore"
)

type config struct {
}

type provider struct {
	Log logs.Logger
	Cfg config

	MySQLXOrm  mysqlxorm.Interface
	EtcdClient *clientv3.Client

	js jsonstore.JsonStore
	types.QueueManager
}

func (p *provider) Init(ctx servicehub.Context) error {
	js, err := jsonstore.New()
	if err != nil {
		return fmt.Errorf("failed to init jsonstore, err: %v", err)
	}
	p.js = js
	//etcdClient, err := etcd.New()
	//if err != nil {
	//	return fmt.Errorf("failed to init etcd client, err: %v", err)
	//}
	p.QueueManager = manager.New(ctx,
		manager.WithDBClient(&dbclient.Client{Engine: p.MySQLXOrm.DB()}),
		//manager.WithEtcdClient(etcdClient),
		manager.WithJsonStore(js),
	)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	go p.continueBackupQueueUsage(ctx)
	go p.QueueManager.ListenInputQueueFromEtcd(ctx)
	go p.QueueManager.ListenUpdatePriorityPipelineIDsFromEtcd(ctx)
	go p.QueueManager.ListenPopOutPipelineIDFromEtcd(ctx)
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("queue-manager", &servicehub.Spec{
		Services:     []string{"queue-manager"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: []string{""},
		Description:  "pipeline engine queue-manager",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
