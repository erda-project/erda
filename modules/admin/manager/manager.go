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

package manager

import (
	"context"
	"net/http"
	"time"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/admin/dao"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type AdminManager struct {
	db        *dao.DBClient
	endpoints []httpserver.Endpoint
	bundle    *bundle.Bundle
	etcdStore *etcd.Store
}

type Option func(am *AdminManager)

func NewAdminManager(options ...Option) *AdminManager {
	admin := &AdminManager{}
	for _, op := range options {
		op(admin)
	}
	return admin
}

func WithDB(db *dao.DBClient) Option {
	return func(am *AdminManager) {
		am.db = db
	}
}

func WithBundle(bundle *bundle.Bundle) Option {
	return func(am *AdminManager) {
		am.bundle = bundle
	}
}

func WithETCDStore(etcdStore *etcd.Store) Option {
	return func(am *AdminManager) {
		am.etcdStore = etcdStore
	}
}

func (am *AdminManager) Routers() []httpserver.Endpoint {
	am.AppendApproveEndpoint()
	am.AppendAuditEndpoint()
	am.AppendNoticeEndpoint()
	am.AppendClusterEndpoint()
	am.AppendHostEndpoint()
	am.AppendAdminEndpoint()
	am.AppendUserEndpoint()
	return am.endpoints
}

func NewBundle() *bundle.Bundle {
	bundleOpts := []bundle.Option{
		bundle.WithCoreServices(),
		bundle.WithClusterManager(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*30),
		)),
	}
	bdl := bundle.New(bundleOpts...)
	return bdl
}

func (am *AdminManager) AppendAdminEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/healthy", Method: http.MethodGet, Handler: am.HealthyCheck},
	}...)
}

func (am *AdminManager) HealthyCheck(
	ctx context.Context, r *http.Request,
	vars map[string]string) (httpserver.Responser, error) {
	return httpserver.OkResp("ok")
}
