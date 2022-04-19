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

package pipelinesvc

import (
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/cache"
	"github.com/erda-project/erda/modules/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/daemon"
	"github.com/erda-project/erda/modules/pipeline/providers/engine"
	"github.com/erda-project/erda/modules/pipeline/providers/run"
	"github.com/erda-project/erda/modules/pipeline/providers/secret"
	"github.com/erda-project/erda/modules/pipeline/providers/user"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/queuemanage"
	"github.com/erda-project/erda/modules/pkg/websocket"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type PipelineSvc struct {
	appSvc          *appsvc.AppSvc
	crondSvc        daemon.Interface
	actionAgentSvc  *actionagentsvc.ActionAgentSvc
	extMarketSvc    *extmarketsvc.ExtMarketSvc
	pipelineCronSvc cronpb.CronServiceServer
	permissionSvc   *permissionsvc.PermissionSvc
	queueManage     *queuemanage.QueueManage
	cache           cache.Interface

	dbClient  *dbclient.Client
	bdl       *bundle.Bundle
	publisher *websocket.Publisher

	engine engine.Interface

	js      jsonstore.JsonStore
	etcdctl *etcd.Store

	// providers
	cmsService  pb.CmsServiceServer
	clusterInfo clusterinfo.Interface
	secret      secret.Interface
	user        user.Interface
	run         run.Interface
}

func New(appSvc *appsvc.AppSvc, crondSvc daemon.Interface,
	actionAgentSvc *actionagentsvc.ActionAgentSvc, extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineCronSvc cronpb.CronServiceServer, permissionSvc *permissionsvc.PermissionSvc,
	queueManage *queuemanage.QueueManage,
	dbClient *dbclient.Client, bdl *bundle.Bundle, publisher *websocket.Publisher,
	engine engine.Interface, js jsonstore.JsonStore, etcd *etcd.Store, clusterInfo clusterinfo.Interface, cache cache.Interface) *PipelineSvc {

	s := PipelineSvc{}
	s.appSvc = appSvc
	s.crondSvc = crondSvc
	s.actionAgentSvc = actionAgentSvc
	s.extMarketSvc = extMarketSvc
	s.pipelineCronSvc = pipelineCronSvc
	s.permissionSvc = permissionSvc
	s.queueManage = queueManage
	s.dbClient = dbClient
	s.bdl = bdl
	s.publisher = publisher
	s.engine = engine
	s.js = js
	s.etcdctl = etcd
	s.clusterInfo = clusterInfo
	s.cache = cache
	return &s
}

func (s *PipelineSvc) WithCmsService(cmsService pb.CmsServiceServer) {
	s.cmsService = cmsService
}

func (s *PipelineSvc) WithSecret(secret secret.Interface) {
	s.secret = secret
}

func (s *PipelineSvc) WithUser(user user.Interface) {
	s.user = user
}

func (s *PipelineSvc) WithRun(run run.Interface) {
	s.run = run
}
