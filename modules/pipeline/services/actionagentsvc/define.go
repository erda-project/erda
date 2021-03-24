package actionagentsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type ActionAgentSvc struct {
	dbClient        *dbclient.Client
	bdl             *bundle.Bundle
	accessibleCache jsonstore.JsonStore
	etcdctl         *etcd.Store
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle, js jsonstore.JsonStore, etcdctl *etcd.Store) *ActionAgentSvc {
	s := ActionAgentSvc{}
	s.dbClient = dbClient
	s.bdl = bdl
	s.accessibleCache = js
	s.etcdctl = etcdctl
	return &s
}
