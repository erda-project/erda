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
	"github.com/erda-project/erda/modules/ecp/dbclient"
	"github.com/erda-project/erda/modules/ecp/services/edge"
	"github.com/erda-project/erda/modules/ecp/services/kubernetes"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type Endpoints struct {
	bdl      *bundle.Bundle
	dbClient *dbclient.DBClient
	edge     *edge.Edge
}

type Option func(*Endpoints)

func New(db *dbclient.DBClient, options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}
	e.dbClient = db
	e.edge = edge.New(
		edge.WithDBClient(db),
		edge.WithBundle(e.bdl),
		edge.WithKubernetes(kubernetes.New()),
	)
	return e
}

// WithBundle With bundle module.
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// Routes Return routes
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/api/edge/site", Method: http.MethodGet, Handler: e.ListEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodGet, Handler: e.GetEdgeSite},
		{Path: "/api/edge/site", Method: http.MethodPost, Handler: e.CreateEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeSite},
		{Path: "/api/edge/site/init/{ID}", Method: http.MethodGet, Handler: e.GetInitEdgeSiteShell},
		{Path: "/api/edge/site/offline/{ID}", Method: http.MethodDelete, Handler: e.OfflineEdgeHost},

		{Path: "/api/edge/configset", Method: http.MethodGet, Handler: e.ListEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodGet, Handler: e.GetEdgeConfigSet},
		{Path: "/api/edge/configset", Method: http.MethodPost, Handler: e.CreateEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeConfigSet},

		{Path: "/api/edge/configset-item", Method: http.MethodGet, Handler: e.ListEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodGet, Handler: e.GetEdgeConfigSetItem},
		{Path: "/api/edge/configset-item", Method: http.MethodPost, Handler: e.CreateEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeConfigSetItem},

		{Path: "/api/edge/app", Method: http.MethodGet, Handler: e.ListEdgeApp},
		{Path: "/api/edge/app", Method: http.MethodPost, Handler: e.CreateEdgeApp},
		{Path: "/api/edge/app/{ID}", Method: http.MethodGet, Handler: e.GetEdgeApp},
		{Path: "/api/edge/app/status/{ID}", Method: http.MethodGet, Handler: e.GetEdgeAppStatus},
		{Path: "/api/edge/app/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeApp},
		{Path: "/api/edge/app/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeApp},

		{Path: "/api/edge/app/site/offline/{ID}", Method: http.MethodPost, Handler: e.OfflineAppSite},
		{Path: "/api/edge/app/site/restart/{ID}", Method: http.MethodPost, Handler: e.RestartAppSite},
	}
}
