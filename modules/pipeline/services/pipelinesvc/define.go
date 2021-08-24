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

package pipelinesvc

import (
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
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
	pluginsManage *plugins_manage.PluginsManage
}

func New(appSvc *appsvc.AppSvc, crondSvc *crondsvc.CrondSvc,
	actionAgentSvc *actionagentsvc.ActionAgentSvc, extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineCronSvc *pipelinecronsvc.PipelineCronSvc, permissionSvc *permissionsvc.PermissionSvc,
	queueManage *queuemanage.QueueManage,
	dbClient *dbclient.Client, bdl *bundle.Bundle, publisher *websocket.Publisher,
	engine *pipengine.Engine, js jsonstore.JsonStore, etcd *etcd.Store,
	pluginsManage *plugins_manage.PluginsManage) *PipelineSvc {

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
	s.pluginsManage = pluginsManage
	return &s
}

func (s *PipelineSvc) WithCmsService(cmsService pb.CmsServiceServer) {
	s.cmsService = cmsService
}
