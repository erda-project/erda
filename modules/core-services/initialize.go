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

// Package core_services
package core_services

import (
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/endpoints"
	"github.com/erda-project/erda/modules/core-services/services/activity"
	"github.com/erda-project/erda/modules/core-services/services/application"
	"github.com/erda-project/erda/modules/core-services/services/approve"
	"github.com/erda-project/erda/modules/core-services/services/audit"
	"github.com/erda-project/erda/modules/core-services/services/errorbox"
	"github.com/erda-project/erda/modules/core-services/services/label"
	"github.com/erda-project/erda/modules/core-services/services/manual_review"
	"github.com/erda-project/erda/modules/core-services/services/mbox"
	"github.com/erda-project/erda/modules/core-services/services/member"
	"github.com/erda-project/erda/modules/core-services/services/notice"
	"github.com/erda-project/erda/modules/core-services/services/notify"
	"github.com/erda-project/erda/modules/core-services/services/org"
	"github.com/erda-project/erda/modules/core-services/services/permission"
	"github.com/erda-project/erda/modules/core-services/services/project"
	"github.com/erda-project/erda/modules/core-services/utils"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/license"
	"github.com/erda-project/erda/pkg/ucauth"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// init endpoints
	ep, err := initEndpoints()
	if err != nil {
		return err
	}

	bdl := bundle.New(
		bundle.WithEventBox(),
		bundle.WithCollector(),
	)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	server.WithLocaleLoader(bdl.GetLocaleLoader())
	// Add auth middleware
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("cmdb"))
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())

	return server.ListenAndServe()
}

// 初始化 Endpoints
func initEndpoints() (*endpoints.Endpoints, error) {
	var (
		etcdStore *etcd.Store
		store     jsonstore.JsonStore
		ossClient *oss.Client
		db        *dao.DBClient
		redisCli  *redis.Client
		err       error
	)

	store, err = jsonstore.New()
	if err != nil {
		return nil, err
	}

	etcdStore, err = etcd.New()
	if err != nil {
		return nil, err
	}

	if utils.IsOSS(conf.AvatarStorageURL()) {
		url, err := url.Parse(conf.AvatarStorageURL())
		if err != nil {
			return nil, err
		}
		appSecret, _ := url.User.Password()
		ossClient, err = oss.New(url.Host, url.User.Username(), appSecret)
		if err != nil {
			return nil, err
		}
	}

	db, err = dao.NewDBClient()
	if err != nil {
		return nil, err
	}

	if conf.LocalMode() {
		redisCli = redis.NewClient(&redis.Options{
			Addr:     conf.RedisAddr(),
			Password: conf.RedisPwd(),
		})
	} else {
		redisCli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    conf.RedisMasterName(),
			SentinelAddrs: strings.Split(conf.RedisSentinelAddrs(), ","),
			Password:      conf.RedisPwd(),
		})
	}
	if _, err := redisCli.Ping().Result(); err != nil {
		return nil, err
	}

	// 初始化UC Client
	uc := ucauth.NewUCClient(discover.UC(), conf.UCClientID(), conf.UCClientSecret())
	if conf.OryEnabled() {
		uc = ucauth.NewUCClient(conf.OryKratosPrivateAddr(), conf.OryCompatibleClientID(), conf.OryCompatibleClientSecret())
	}

	// init bundle
	bundleOpts := []bundle.Option{
		bundle.WithAddOnPlatform(),
		bundle.WithGittar(),
		bundle.WithGittarAdaptor(),
		bundle.WithEventBox(),
		bundle.WithMonitor(),
		bundle.WithScheduler(),
		bundle.WithDiceHub(),
		bundle.WithPipeline(),
		bundle.WithOrchestrator(),
		bundle.WithQA(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*30),
		)),
		bundle.WithKMS(),
		bundle.WithHepa(),
		bundle.WithCollector(),
	}
	bdl := bundle.New(bundleOpts...)

	// init org service
	o := org.New(
		org.WithDBClient(db),
		org.WithUCClient(uc),
		org.WithBundle(bdl),
		org.WithRedisClient(redisCli),
	)

	// init project service
	p := project.New(
		project.WithDBClient(db),
		project.WithUCClient(uc),
		project.WithBundle(bdl),
	)

	// init app service
	app := application.New(
		application.WithDBClient(db),
		application.WithUCClient(uc),
		application.WithBundle(bdl),
	)

	// init member service
	m := member.New(
		member.WithDBClient(db),
		member.WithUCClient(uc),
		member.WithRedisClient(redisCli),
	)
	mr := manual_review.New(
		manual_review.WithDBClient(db),
	)
	notifyService := notify.New(
		notify.WithDBClient(db),
	)

	// init activity service
	a := activity.New(
		activity.WithDBClient(db),
		activity.WithBundle(bdl),
	)

	pm := permission.New(
		permission.WithDBClient(db),
	)

	mboxService := mbox.New(
		mbox.WithDBClient(db),
		mbox.WithBundle(bdl),
	)

	approve := approve.New(
		approve.WithDBClient(db),
		approve.WithBundle(bdl),
		approve.WithUCClient(uc),
		approve.WithMember(m),
	)

	//通过ui显示错误,不影响启动
	license, _ := license.ParseLicense(conf.LicenseKey())

	// init label
	l := label.New(
		label.WithDBClient(db),
	)

	notice := notice.New(
		notice.WithDBClient(db),
	)

	audit := audit.New(
		audit.WithDBClient(db),
		audit.WithUCClient(uc),
	)

	errorBox := errorbox.New(
		errorbox.WithDBClient(db),
		errorbox.WithBundle(bdl),
	)

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithJSONStore(store),
		endpoints.WithEtcdStore(etcdStore),
		endpoints.WithOSSClient(ossClient),
		endpoints.WithDBClient(db),
		endpoints.WithUCClient(uc),
		endpoints.WithBundle(bdl),
		endpoints.WithOrg(o),
		endpoints.WithManualReview(mr),
		endpoints.WithProject(p),
		endpoints.WithApp(app),
		endpoints.WithMember(m),
		endpoints.WithActivity(a),
		endpoints.WithPermission(pm),
		endpoints.WithNotify(notifyService),
		endpoints.WithLicense(license),
		endpoints.WithLabel(l),
		endpoints.WithMBox(mboxService),
		endpoints.WithNotice(notice),
		endpoints.WithApprove(approve),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithAudit(audit),
		endpoints.WithErrorBox(errorBox),
	)

	return ep, nil
}
