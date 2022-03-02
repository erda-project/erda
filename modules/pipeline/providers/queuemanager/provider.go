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

package queuemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/dispatcher"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanager/manager"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanager/types"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type provider struct {
	Log logs.Logger
	Cfg *config

	MySQLXOrm  mysqlxorm.Interface
	EtcdClient *clientv3.Client
	Lw         leaderworker.Interface
	Dispatcher dispatcher.Interface

	dbClient *dbclient.Client
	js       jsonstore.JsonStore
	types.QueueManager
}

func (q *provider) Init(ctx servicehub.Context) error {
	if len(q.Cfg.IncomingPipelineCfg.EtcdKeyPrefixWithSlash) == 0 {
		return fmt.Errorf("failed to find config: incoming_pipeline.etcd_key_prefix_with_slash")
	}
	q.Cfg.IncomingPipelineCfg.EtcdKeyPrefixWithSlash = filepath.Clean(q.Cfg.IncomingPipelineCfg.EtcdKeyPrefixWithSlash) + "/"

	js, err := jsonstore.New()
	if err != nil {
		return fmt.Errorf("failed to init jsonstore, err: %v", err)
	}
	q.js = js
	etcdClient, err := etcd.New()
	if err != nil {
		return fmt.Errorf("failed to init etcd client, err: %v", err)
	}
	q.dbClient = &dbclient.Client{Engine: q.MySQLXOrm.DB()}
	q.QueueManager = manager.New(ctx,
		manager.WithDBClient(&dbclient.Client{Engine: q.MySQLXOrm.DB()}),
		manager.WithEtcdClient(etcdClient),
		manager.WithJsonStore(js),
	)
	return nil
}

func (q *provider) Run(ctx context.Context) error {
	q.Lw.OnLeader(q.continueBackupQueueUsage)
	q.Lw.OnLeader(q.QueueManager.ListenInputQueueFromEtcd)
	q.Lw.OnLeader(q.QueueManager.ListenUpdatePriorityPipelineIDsFromEtcd)
	q.Lw.OnLeader(q.QueueManager.ListenPopOutPipelineIDFromEtcd)
	q.Lw.OnLeader(q.listenIncomingPipeline)
	q.Lw.OnLeader(q.loadRunningPipelines)
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
