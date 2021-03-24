package pipelinesvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/cmsvc"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinecronsvc"
	"github.com/erda-project/erda/modules/pkg/websocket"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type PipelineSvc struct {
	appSvc          *appsvc.AppSvc
	cmSvc           *cmsvc.CMSvc
	crondSvc        *crondsvc.CrondSvc
	actionAgentSvc  *actionagentsvc.ActionAgentSvc
	extMarketSvc    *extmarketsvc.ExtMarketSvc
	pipelineCronSvc *pipelinecronsvc.PipelineCronSvc
	permissionSvc   *permissionsvc.PermissionSvc

	dbClient  *dbclient.Client
	bdl       *bundle.Bundle
	publisher *websocket.Publisher

	engine *pipengine.Engine

	js      jsonstore.JsonStore
	etcdctl *etcd.Store
}

func New(appSvc *appsvc.AppSvc, cmSvc *cmsvc.CMSvc, crondSvc *crondsvc.CrondSvc,
	actionAgentSvc *actionagentsvc.ActionAgentSvc, extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineCronSvc *pipelinecronsvc.PipelineCronSvc, permissionSvc *permissionsvc.PermissionSvc,
	dbClient *dbclient.Client, bdl *bundle.Bundle, publisher *websocket.Publisher,
	engine *pipengine.Engine, js jsonstore.JsonStore, etcd *etcd.Store) *PipelineSvc {

	s := PipelineSvc{}
	s.appSvc = appSvc
	s.cmSvc = cmSvc
	s.crondSvc = crondSvc
	s.actionAgentSvc = actionAgentSvc
	s.extMarketSvc = extMarketSvc
	s.pipelineCronSvc = pipelineCronSvc
	s.permissionSvc = permissionSvc
	s.dbClient = dbClient
	s.bdl = bdl
	s.publisher = publisher
	s.engine = engine
	s.js = js
	s.etcdctl = etcd
	return &s
}
