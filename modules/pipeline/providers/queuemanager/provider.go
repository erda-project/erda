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
	"time"

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

type config struct {
	IncomingPipelineCfg IncomingPipelineCfg `file:"incoming_pipeline"`
}

type IncomingPipelineCfg struct {
	ListenPrefixWithSlash string        `file:"listen_prefix" default:"/devops/pipeline/v2/queue-manager/incoming_pipeline/"`
	RetryInterval         time.Duration `file:"retry_interval" default:"10s"`
}

type provider struct {
	Log logs.Logger
	Cfg config

	MySQLXOrm  mysqlxorm.Interface
	EtcdClient *clientv3.Client
	Lw         leaderworker.Interface
	Dispatcher dispatcher.Interface

	dbClient *dbclient.Client
	js       jsonstore.JsonStore
	types.QueueManager
}

func (q *provider) Init(ctx servicehub.Context) error {
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
	q.Cfg.IncomingPipelineCfg.ListenPrefixWithSlash = filepath.Clean(q.Cfg.IncomingPipelineCfg.ListenPrefixWithSlash) + "/"
	return nil
}

func (q *provider) Run(ctx context.Context) error {
	q.Lw.OnLeader(q.continueBackupQueueUsage)
	q.Lw.OnLeader(q.QueueManager.ListenInputQueueFromEtcd)
	q.Lw.OnLeader(q.QueueManager.ListenUpdatePriorityPipelineIDsFromEtcd)
	q.Lw.OnLeader(q.QueueManager.ListenPopOutPipelineIDFromEtcd)
	q.Lw.OnLeader(q.listenIncomingPipeline)
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
