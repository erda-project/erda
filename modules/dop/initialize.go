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

package dop

import (
	"time"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/endpoints"
	"github.com/erda-project/erda/modules/dop/event"
	"github.com/erda-project/erda/modules/dop/services/apidocsvc"
	"github.com/erda-project/erda/modules/dop/services/assetsvc"
	"github.com/erda-project/erda/modules/dop/services/autotest"
	atv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"github.com/erda-project/erda/modules/dop/services/cdp"
	"github.com/erda-project/erda/modules/dop/services/cq"
	"github.com/erda-project/erda/modules/dop/services/filetree"
	"github.com/erda-project/erda/modules/dop/services/migrate"
	"github.com/erda-project/erda/modules/dop/services/permission"
	"github.com/erda-project/erda/modules/dop/services/pipeline"
	"github.com/erda-project/erda/modules/dop/services/sceneset"
	"github.com/erda-project/erda/modules/dop/services/sonar_metric_rule"
	"github.com/erda-project/erda/modules/dop/services/testcase"
	"github.com/erda-project/erda/modules/dop/services/testplan"
	"github.com/erda-project/erda/modules/dop/services/testset"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("set log level: %s", logrus.DebugLevel)
	}

	// TODO invoke self use service
	//_ = os.Setenv("QA_ADDR", discover.QA())

	// init db
	if err := dbclient.Open(); err != nil {
		return err
	}
	defer dbclient.Close()

	ep := initEndpoints((*dao.DBClient)(dbclient.DB))

	// 注册 hook
	if err := ep.RegisterEvents(); err != nil {
		return err
	}

	server := httpserver.New(conf.ListenAddr())
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())
	server.Router().PathPrefix("/api/apim/metrics").Handler(endpoints.InternalReverseHandler(endpoints.ProxyMetrics))

	loadMetricKeysFromDb((*dao.DBClient)(dbclient.DB))
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())
	return server.ListenAndServe()
}

func initEndpoints(db *dao.DBClient) *endpoints.Endpoints {
	// init bundle
	bdl.Init(
		bundle.WithCMDB(),
		bundle.WithHepa(),
		bundle.WithOrchestrator(),
		bundle.WithEventBox(),
		bundle.WithGittar(),
		bundle.WithPipeline(),
		bundle.WithMonitor(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second*15, time.Duration(conf.BundleTimeoutSecond())*time.Second), // bundle 默认 (time.Second, time.Second*3)
		)),
	)

	// init pipeline
	p := pipeline.New(pipeline.WithBundle(bdl.Bdl))

	c := cdp.New(cdp.WithBundle(bdl.Bdl))

	// init event
	e := event.New(event.WithBundle(bdl.Bdl))

	// init permission
	perm := permission.New(permission.WithBundle(bdl.Bdl))

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	fileTree := filetree.New(filetree.WithBundle(bdl.Bdl))

	// init service
	assetSvc := assetsvc.New()
	filetreeSvc := apidocsvc.New()

	testCaseSvc := testcase.New(
		testcase.WithDBClient(db),
		testcase.WithBundle(bdl.Bdl),
	)
	testSetSvc := testset.New(
		testset.WithDBClient(db),
		testset.WithBundle(bdl.Bdl),
		testset.WithTestCaseService(testCaseSvc),
	)

	testCaseSvc.CreateTestSetFn = testSetSvc.Create

	autotest := autotest.New(autotest.WithDBClient(db), autotest.WithBundle(bdl.Bdl))

	sceneset := sceneset.New(
		sceneset.WithDBClient(db),
		sceneset.WithBundle(bdl.Bdl),
	)

	autotestV2 := atv2.New(atv2.WithDBClient(db), atv2.WithBundle(bdl.Bdl), atv2.WithSceneSet(sceneset))

	sceneset.GetScenes = autotestV2.ListAutotestScene
	sceneset.CopyScene = autotestV2.CopyAutotestScene

	testPlan := testplan.New(
		testplan.WithDBClient(db),
		testplan.WithBundle(bdl.Bdl),
		testplan.WithTestCase(testCaseSvc),
		testplan.WithTestSet(testSetSvc),
		testplan.WithAutoTest(autotest),
	)

	sonarMetricRule := sonar_metric_rule.New(
		sonar_metric_rule.WithDBClient(db),
		sonar_metric_rule.WithBundle(bdl.Bdl),
	)

	migrateSvc := migrate.New(migrate.WithDBClient(db))

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithBundle(bdl.Bdl),
		endpoints.WithPipeline(p),
		endpoints.WithEvent(e),
		endpoints.WithCDP(c),
		endpoints.WithPermission(perm),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithGittarFileTree(fileTree),

		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithAssetSvc(assetSvc),
		endpoints.WithFileTreeSvc(filetreeSvc),

		endpoints.WithDB(db),
		endpoints.WithTestcase(testCaseSvc),
		endpoints.WithTestSet(testSetSvc),
		endpoints.WithSonarMetricRule(sonarMetricRule),
		endpoints.WithTestplan(testPlan),
		endpoints.WithCQ(cq.New(cq.WithBundle(bdl.Bdl))),
		endpoints.WithAutoTest(autotest),
		endpoints.WithAutoTestV2(autotestV2),
		endpoints.WithSceneSet(sceneset),
		endpoints.WithMigrate(migrateSvc),
	)

	return ep
}

func loadMetricKeysFromDb(db *dao.DBClient) {
	var list []*apistructs.SonarMetricKey
	if err := db.Table("qa_sonar_metric_keys").Find(&list).Error; err != nil {
		panic(err)
	}

	for _, sonarMetricKey := range list {
		apistructs.SonarMetricKeys[sonarMetricKey.ID] = sonarMetricKey
	}
}
