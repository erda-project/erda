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

// Package cmdb dice元数据管理
package cmdb

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/conf"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/endpoints"
	"github.com/erda-project/erda/modules/cmdb/services/cloudaccount"
	"github.com/erda-project/erda/modules/cmdb/services/cluster"
	"github.com/erda-project/erda/modules/cmdb/services/container"
	"github.com/erda-project/erda/modules/cmdb/services/host"
	"github.com/erda-project/erda/modules/cmdb/services/org"
	"github.com/erda-project/erda/modules/cmdb/utils"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/license"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
	// "terminus.io/dice/telemetry/promxp"
)

// 数据库表 cm_container gc 的周期
const containerGCPeriod = 3 * 24 * time.Hour

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

	// TODO:
	// 启动消费者协程，用于消费 kafka 消息
	// go func() {
	// 	logrus.Info("start Consumer....")
	// 	ep.Consumer()
	// }()

	go runContainerGC(ep.DBClient())

	go func() {
		dao.Count()
	}()

	// 定时同步主机实际资源使用值至DB
	go func() {
		ep.SyncHostResource(conf.HostSyncInterval())
	}()

	// 定时同步任务(job/deployment)状态至DB
	go func() {
		ep.SyncTaskStatus(conf.TaskSyncDuration())
	}()

	// 定时清理任务(job/deployment)信息
	go func() {
		ep.TaskClean(conf.TaskCleanDuration())
	}()

	bdl := bundle.New(
		bundle.WithEventBox(),
		bundle.WithCollector(),
	)

	registerWebHook(bdl)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	server.WithLocaleLoader(bdl.GetLocaleLoader())
	// Add auth middleware
	server.Router().Use(authenticate)
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("cmdb"))
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())
	logrus.Errorf("[alert] starting cmdb instance...")

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

	// init account service
	account := cloudaccount.New(
		cloudaccount.WithDBClient(db),
	)

	con := container.New(
		container.WithDBClient(db),
		container.WithBundle(bdl),
	)

	h := host.New(
		host.WithDBClient(db),
		host.WithBundle(bdl),
		host.WithContainer(con),
	)

	cl := cluster.New(
		cluster.WithDBClient(db),
		cluster.WithHostService(h),
		cluster.WithBundle(bdl),
		cluster.WithContainerService(con),
	)

	//通过ui显示错误,不影响启动
	license, _ := license.ParseLicense(conf.LicenseKey())

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
		endpoints.WithCloudAccount(account),
		endpoints.WithHost(h),
		endpoints.WithContainer(con),
		endpoints.WithCluster(cl),
		endpoints.WithLicense(license),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
	)

	return ep, nil
}

// 清理 cm_containers 数据库表的过期数据
func runContainerGC(db *dao.DBClient) {
	logrus.Info("Start to run container gc.")
	defer logrus.Info("End running container gc.")

	// 1h 运行一次清理
	loopTime := time.Hour
	ticker := time.NewTicker(loopTime)
	defer ticker.Stop()

	for {
		if err := db.DeleteStoppedContainersByPeriod(context.Background(), containerGCPeriod); err != nil {
			logrus.Printf("failed to delete all stopped containers: %v", err)
		}
		<-ticker.C
	}
}

// API认证
func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// authentication
		authenticated, err := Authentication(r)
		if err != nil {
			logrus.Printf("authentication error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authenticated {
			logrus.Println("authentication failed !!")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authentication(r *http.Request) (bool, error) {
	if r == nil {
		return false, errors.Errorf("invalid param: request is nil")
	}

	// 特权模式
	previleged := r.Header.Get("Previleged")
	if previleged == "true" {
		logrus.Infof("%s %s: previleged is true", r.Method, r.URL.String())
		return true, nil
	}

	// Client-ID & Client-Name
	// 如果使用 token 的方式，先跳过鉴权
	clientID := r.Header.Get("Client-ID")
	if clientID != "" {
		logrus.Infof("%s %s: token mode", r.Method, r.URL.String())
		return true, nil
	}

	return true, nil
}

func registerWebHook(bdl *bundle.Bundle) {
	// register pipeline tasks by webhook
	ev := apistructs.CreateHookRequest{
		Name:   "cmdb_pipeline_tasks",
		Events: []string{"pipeline_task", "pipeline_task_runtime"},
		URL:    strutil.Concat("http://", discover.CMDB(), "/api/tasks"),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register pipeline tasks event, (%v)", err)
	}

	// 注册审批流状态变更监听
	ev = apistructs.CreateHookRequest{
		Name:   "cmdb_approve_status_changed",
		Events: []string{bundle.ApprovalStatusChangedEvent},
		URL:    strutil.Concat("http://", discover.CMDB(), "/api/approvals/actions/watch-status"),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register approval status changed event, %v", err)
	}
}
