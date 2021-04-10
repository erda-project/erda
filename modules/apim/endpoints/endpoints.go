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

// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"context"
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/pkg/httpserver"

	"github.com/erda-project/erda/modules/apim/services/apidocsvc"
	"github.com/erda-project/erda/modules/apim/services/assetsvc"
)

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/health", Method: http.MethodGet, Handler: e.Health},

		{Path: "/api/api-assets", Method: http.MethodPost, Handler: e.CreateAPIAsset},
		{Path: "/api/api-assets", Method: http.MethodGet, Handler: e.PagingAPIAssets},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodGet, Handler: e.GetAPIAsset},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodPut, Handler: e.UpdateAPIAsset},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodDelete, Handler: e.DeleteAPIAsset},

		{Path: "/api/api-assets/{assetID}/api-gateways", Method: http.MethodGet, Handler: e.ListAPIGateways},
		{Path: "/api/api-gateways/{projectID}", Method: http.MethodGet, Handler: e.ListProjectAPIGateways},

		{Path: "/api/api-assets/{assetID}/versions", Method: http.MethodGet, Handler: e.PagingAPIAssetVersions},
		{Path: "/api/api-assets/{assetID}/versions", Method: http.MethodPost, Handler: e.CreateAPIVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodGet, Handler: e.GetAPIAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodPut, Handler: e.UpdateAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodDelete, Handler: e.DeleteAPIAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}/export", Method: http.MethodGet, WriterHandler: e.DownloadSpecText},

		{Path: "/api/api-assets/{assetID}/swagger-versions", Method: http.MethodGet, Handler: e.ListSwaggerVersions},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/clients", Method: http.MethodGet, Handler: e.ListSwaggerClient},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/attempt-test", Method: http.MethodPost, Handler: e.ExecuteAttemptTest},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations", Method: http.MethodPost, Handler: e.CreateInstantiation},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations", Method: http.MethodGet, Handler: e.GetInstantiations},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations/{instantiationID}", Method: http.MethodPut, Handler: e.UpdateInstantiation},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas", Method: http.MethodGet, Handler: e.ListSLAs},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas", Method: http.MethodPost, Handler: e.CreateSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodGet, Handler: e.GetSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodDelete, Handler: e.DeleteSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodPut, Handler: e.UpdateSLA},

		{Path: "/api/api-clients", Method: http.MethodPost, Handler: e.CreateClient},
		{Path: "/api/api-clients", Method: http.MethodGet, Handler: e.ListMyClients},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodGet, Handler: e.GetClient},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodPut, Handler: e.UpdateClient},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodDelete, Handler: e.DeleteClient},

		{Path: "/api/api-clients/{clientID}/contracts", Method: http.MethodPost, Handler: e.CreateContract},
		{Path: "/api/api-clients/{clientID}/contracts", Method: http.MethodGet, Handler: e.ListContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodGet, Handler: e.GetContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodPut, Handler: e.UpdateContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodDelete, Handler: e.DeleteContract},

		{Path: "/api/api-clients/{clientID}/contracts/{contractID}/operation-records", Method: http.MethodGet, Handler: e.ListContractRecords},

		{Path: "/api/api-access", Method: http.MethodPost, Handler: e.CreateAccess},
		{Path: "/api/api-access", Method: http.MethodGet, Handler: e.ListAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodGet, Handler: e.GetAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodPut, Handler: e.UpdateAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodDelete, Handler: e.DeleteAccess},

		{Path: "/api/api-app-services/{appID}", Method: http.MethodGet, Handler: e.ListRuntimeServices},

		{Path: "/api/apim-ws/api-docs/filetree/{inode}", Method: http.MethodGet, WriterHandler: e.APIDocWebsocket},
		{Path: "/api/apim/{treeName}/filetree", Method: http.MethodPost, Handler: e.CreateNode},
		{Path: "/api/apim/{treeName}/filetree", Method: http.MethodGet, Handler: e.ListChildrenNodes},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodDelete, Handler: e.DeleteNode},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodPut, Handler: e.UpdateNode},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodGet, Handler: e.GetNodeDetail},
		{Path: "/api/apim/{treeName}/filetree/{inode}/actions/{action}", Method: http.MethodPost, Handler: e.MvCpNode},

		{Path: "/api/apim/operations", Method: http.MethodGet, Handler: e.SearchOperations},
		{Path: "/api/apim/operations/{id}", Method: http.MethodGet, Handler: e.GetOperation},

		{Path: "/api/apim/validate-swagger", Method: http.MethodPost, Handler: e.ValidateSwagger},
	}
}

func NotImplemented(ctx context.Context, request *http.Request, m map[string]string) (httpserver.Responser, error) {
	return httpserver.ErrResp(http.StatusNotImplemented, "", "not implemented")
}

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	queryStringDecoder *schema.Decoder

	assetSvc    *assetsvc.Service
	fileTreeSvc *apidocsvc.Service
}

type Option func(*Endpoints)

func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithAssetSvc(svc *assetsvc.Service) Option {
	return func(e *Endpoints) {
		e.assetSvc = svc
	}
}

func WithFileTreeSvc(svc *apidocsvc.Service) Option {
	return func(e *Endpoints) {
		e.fileTreeSvc = svc
	}
}
