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

package endpoints

import (
	"net/http"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cluster-manager/dbclient"
	"github.com/erda-project/erda/modules/cluster-manager/services/cluster"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type Endpoints struct {
	dbClient *dbclient.DBClient
	cluster  *cluster.Cluster
}

type Option func(*Endpoints)

func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	e.cluster = cluster.New(
		cluster.WithDBClient(e.dbClient),
		cluster.WithBundle(bundle.New(bundle.WithEventBox())),
	)

	return e
}

func WithDBClient(db *dbclient.DBClient) Option {
	return func(e *Endpoints) {
		e.dbClient = db
	}
}

// Routes Return routes
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/api/clusters", Method: http.MethodGet, Handler: auth(e.ListCluster)},
		{Path: "/api/clusters/{idOrName}", Method: http.MethodGet, Handler: auth(e.GetCluster)},
		{Path: "/api/clusters", Method: http.MethodPost, Handler: auth(e.CreateCluster)},
		{Path: "/api/clusters", Method: http.MethodPut, Handler: auth(e.UpdateCluster)},
		{Path: "/api/clusters/{clusterName}", Method: http.MethodDelete, Handler: auth(e.DeleteCluster)},
		{Path: "/api/clusters", Method: http.MethodPatch, Handler: auth(e.PatchCluster)},
	}
}
