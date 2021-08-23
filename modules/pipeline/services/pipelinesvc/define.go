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
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinecronsvc"
	"github.com/erda-project/erda/modules/pipeline/services/queuemanage"
	"github.com/erda-project/erda/modules/pkg/websocket"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type PipelineSvc struct {
	appSvc          *appsvc.AppSvc
	crondSvc        *crondsvc.CrondSvc
	actionAgentSvc  *actionagentsvc.ActionAgentSvc
	extMarketSvc    *extmarketsvc.ExtMarketSvc
	pipelineCronSvc *pipelinecronsvc.PipelineCronSvc
	permissionSvc   *permissionsvc.PermissionSvc
	queueManage     *queuemanage.QueueManage

	dbClient  *dbclient.Client
	bdl       *bundle.Bundle
	publisher *websocket.Publisher

	engine *pipengine.Engine

	js      jsonstore.JsonStore
	etcdctl *etcd.Store

	// providers
	cmsService pb.CmsServiceServer
}

func New(appSvc *appsvc.AppSvc, crondSvc *crondsvc.CrondSvc,
	actionAgentSvc *actionagentsvc.ActionAgentSvc, extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineCronSvc *pipelinecronsvc.PipelineCronSvc, permissionSvc *permissionsvc.PermissionSvc,
	queueManage *queuemanage.QueueManage,
	dbClient *dbclient.Client, bdl *bundle.Bundle, publisher *websocket.Publisher,
	engine *pipengine.Engine, js jsonstore.JsonStore, etcd *etcd.Store) *PipelineSvc {

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
	return &s
}

func (s *PipelineSvc) WithCmsService(cmsService pb.CmsServiceServer) {
	s.cmsService = cmsService
}
