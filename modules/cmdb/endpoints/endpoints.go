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
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gorilla/schema"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/cloudaccount"
	"github.com/erda-project/erda/modules/cmdb/services/cluster"
	"github.com/erda-project/erda/modules/cmdb/services/container"
	"github.com/erda-project/erda/modules/cmdb/services/host"
	"github.com/erda-project/erda/modules/cmdb/services/org"
	"github.com/erda-project/erda/modules/core-services/services/errorbox"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/license"
	"github.com/erda-project/erda/pkg/ucauth"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	store              jsonstore.JsonStore
	etcdStore          *etcd.Store
	ossClient          *oss.Client
	db                 *dao.DBClient
	uc                 *ucauth.UCClient
	bdl                *bundle.Bundle
	org                *org.Org
	cloudaccount       *cloudaccount.CloudAccount
	host               *host.Host
	container          *container.Container
	cluster            *cluster.Cluster
	license            *license.License
	queryStringDecoder *schema.Decoder
	errorbox           *errorbox.ErrorBox
}

type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

// WithOSSClient 配置OSS Client
func WithOSSClient(client *oss.Client) Option {
	return func(e *Endpoints) {
		e.ossClient = client
	}
}

// WithDBClient 配置 db
func WithDBClient(db *dao.DBClient) Option {
	return func(e *Endpoints) {
		e.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// WithUCClient 配置 UC Client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(e *Endpoints) {
		e.uc = uc
	}
}

// WithJSONStore 配置 jsonstore
func WithJSONStore(store jsonstore.JsonStore) Option {
	return func(e *Endpoints) {
		e.store = store
	}
}

// WithEtcdStore 配置 etcdStore
func WithEtcdStore(etcdStore *etcd.Store) Option {
	return func(e *Endpoints) {
		e.etcdStore = etcdStore
	}
}

// WithOrg 配置 org service
func WithOrg(org *org.Org) Option {
	return func(e *Endpoints) {
		e.org = org
	}
}

// WithCloudAccount 配置 cloudaccount service
func WithCloudAccount(account *cloudaccount.CloudAccount) Option {
	return func(e *Endpoints) {
		e.cloudaccount = account
	}
}

// WithHost 配置 host service
func WithHost(host *host.Host) Option {
	return func(e *Endpoints) {
		e.host = host
	}
}

// WithContainer 配置 container service
func WithContainer(container *container.Container) Option {
	return func(e *Endpoints) {
		e.container = container
	}
}

// WithCluster 配置 cluster service
func WithCluster(cluster *cluster.Cluster) Option {
	return func(e *Endpoints) {
		e.cluster = cluster
	}
}

// WithLicense 配置 license
func WithLicense(license *license.License) Option {
	return func(e *Endpoints) {
		e.license = license
	}
}

// WithQueryStringDecoder 配置 queryStringDecoder
func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithErrorBox(errorbox *errorbox.ErrorBox) Option {
	return func(e *Endpoints) {
		e.errorbox = errorbox
	}
}

// DBClient 获取db client
func (e *Endpoints) DBClient() *dao.DBClient {
	return e.db
}

func (e *Endpoints) UCClient() *ucauth.UCClient {
	return e.uc
}

// GetLocale 获取本地化资源
func (e *Endpoints) GetLocale(request *http.Request) *i18n.LocaleResource {
	return e.bdl.GetLocaleByRequest(request)
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		// hosts
		{Path: "/api/hosts/{host}", Method: http.MethodGet, Handler: e.GetHost},
		{Path: "/api/org/actions/list-running-tasks", Method: http.MethodGet, Handler: e.ListOrgRunningTasks},
		{Path: "/api/tasks", Method: http.MethodPost, Handler: e.DealTaskEvent},

		// 仅供监控使用，不在 openapi 暴露
		{Path: "/api/containers/actions/list-edas", Method: http.MethodGet, Handler: e.ListEdasContainers},

		// webhook
		{Path: "/api/events/instance-status", Method: http.MethodPost, Handler: e.UpdateInstanceBySchedulerEvent},

		// 集群相关
		{Path: "/api/clusters", Method: http.MethodPost, Handler: e.CreateCluster},
		{Path: "/api/clusters", Method: http.MethodPut, Handler: e.UpdateCluster},
		{Path: "/api/clusters/{idOrName}", Method: http.MethodGet, Handler: e.GetCluster},
		{Path: "/api/clusters", Method: http.MethodGet, Handler: e.ListCluster},
		{Path: "/api/clusters/{clusterName}", Method: http.MethodDelete, Handler: e.DeleteCluster},
		{Path: "/api/clusters/actions/dereference", Method: http.MethodPut, Handler: e.DereferenceCluster},

		// 云账号相关
		{Path: "/api/cloud-accounts", Method: http.MethodPost, Handler: e.CreateCloudAccount},
		{Path: "/api/cloud-accounts", Method: http.MethodGet, Handler: e.ListCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodGet, Handler: e.GetCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodPut, Handler: e.UpdateCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodDelete, Handler: e.DeleteCloudAccount},

		// 用户相关
		{Path: "/api/users", Method: http.MethodGet, Handler: e.ListUser},
		{Path: "/api/users/current", Method: http.MethodGet, Handler: e.GetCurrentUser},
		{Path: "/api/users/actions/search", Method: http.MethodGet, Handler: e.SearchUser},
	}
}
