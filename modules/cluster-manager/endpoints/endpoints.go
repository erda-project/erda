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
