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
