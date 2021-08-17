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

// Package dicehub Dicehub module
package dicehub

import (
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/conf"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/endpoints"
	"github.com/erda-project/erda/modules/dicehub/recycle"
	"github.com/erda-project/erda/modules/dicehub/service/extension"
	"github.com/erda-project/erda/modules/dicehub/service/publish_item"
	"github.com/erda-project/erda/modules/dicehub/service/release"
	"github.com/erda-project/erda/modules/dicehub/service/template"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize(p *provider) error {
	// 加载环境变量配置
	conf.Load()
	// 设置日志级别
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// init endpoints
	ep, err := initEndpoints(p)
	if err != nil {
		return err
	}

	// extension init
	go InitExtension(ep.Extension())

	// 启动 release 定时清理任务
	if err := ReleaseGC(ep.Release()); err != nil {
		return err
	}

	server := httpserver.New(conf.ListenAddr())
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("dicehub"))
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())

	// TODO pull操作用于获取diceyml不合适，需预留给更通用的操作,为了保持兼容，暂时先保留,推动业务方使用get-dice API
	server.Router().Path("/api/releases/{releaseId}/actions/pull").Methods(http.MethodGet).HandlerFunc(ep.GetDiceYAML)
	server.Router().Path("/api/releases/{releaseId}/actions/get-dice").Methods(http.MethodGet).HandlerFunc(ep.GetDiceYAML)

	return server.ListenAndServe()
}

// ReleaseGC 启动release gc定时任务
func ReleaseGC(rl *release.Release) error {
	if conf.GCSwitch() {
		etcdStore, err := etcd.New()
		if err != nil {
			logrus.Errorf("[alert] initialize etcd client error: %v", err)
			return err
		}
		recycle.ImageGCCron(rl, etcdStore.GetClient())
	}
	return nil
}

func InitExtension(ex *extension.Extension) {
	err := ex.InitExtension("/app/extensions")
	if err != nil {
		panic(err)
	}
	logrus.Infoln("End init extension")
}

// 初始化 Endpoints
func initEndpoints(p *provider) (*endpoints.Endpoints, error) {
	// 数据库初始化
	db, err := dbclient.Open()
	if err != nil {
		return nil, err
	}

	// init bundle
	bundleOpts := []bundle.Option{
		bundle.WithEventBox(),
		bundle.WithCoreServices(),
		bundle.WithDOP(),
		bundle.WithCMP(),
		bundle.WithMonitor(),
		bundle.WithPipeline(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)
	rl := release.New(
		release.WithDBClient(db),
		release.WithBundle(bdl),
		release.WithImageDBClient(p.ImageDB),
	)

	ext := extension.New(
		extension.WithDBClient(db),
		extension.WithBundle(bdl),
	)

	publishItem := publish_item.New(
		publish_item.WithDBClient(db),
		publish_item.WithBundle(bdl),
	)
	// 3.20 灰度逻辑迁移，3.21删除
	// publishItem.Migration320()

	pipelineTemplate := template.New(
		template.WithBundle(bdl),
		template.WithDBClient(db),
	)

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(db),
		endpoints.WithBundle(bdl),
		endpoints.WithRelease(rl),
		endpoints.WithExtension(ext),
		endpoints.WithPublishItem(publishItem),
		endpoints.WithPipelineTemplate(pipelineTemplate),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
	)

	return ep, nil
}
