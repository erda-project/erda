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

package qa

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/qa/conf"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/endpoints"
	"github.com/erda-project/erda/modules/qa/services/autotest"
	atv2 "github.com/erda-project/erda/modules/qa/services/autotest_v2"
	"github.com/erda-project/erda/modules/qa/services/cq"
	"github.com/erda-project/erda/modules/qa/services/migrate"
	"github.com/erda-project/erda/modules/qa/services/sceneset"
	"github.com/erda-project/erda/modules/qa/services/sonar_metric_rule"
	"github.com/erda-project/erda/modules/qa/services/testcase"
	"github.com/erda-project/erda/modules/qa/services/testplan"
	"github.com/erda-project/erda/modules/qa/services/testset"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
)

// Initialize qa server initialize
func Initialize() error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// TODO invoke self use service
	_ = os.Setenv("QA_ADDR", conf.SelfAddr())

	// init bundle
	bdl := bundle.New(
		bundle.WithQA(),
		bundle.WithPipeline(),
		bundle.WithCMDB(),
		bundle.WithMonitor(),
		bundle.WithGittar(),
		bundle.WithEventBox(),
		bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second*5, time.Second*30))),
	)

	// init query string decoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// init db
	db, err := dao.Open()
	if err != nil {
		return err
	}

	testCaseSvc := testcase.New(
		testcase.WithDBClient(db),
		testcase.WithBundle(bdl),
	)

	testSetSvc := testset.New(
		testset.WithDBClient(db),
		testset.WithBundle(bdl),
		testset.WithTestCaseService(testCaseSvc),
	)

	testCaseSvc.CreateTestSetFn = testSetSvc.Create

	autotest := autotest.New(autotest.WithDBClient(db), autotest.WithBundle(bdl))

	sceneset := sceneset.New(
		sceneset.WithDBClient(db),
		sceneset.WithBundle(bdl),
	)

	autotestV2 := atv2.New(atv2.WithDBClient(db), atv2.WithBundle(bdl), atv2.WithSceneSet(sceneset))

	sceneset.GetScenes = autotestV2.ListAutotestScene
	sceneset.CopyScene = autotestV2.CopyAutotestScene

	testPlan := testplan.New(
		testplan.WithDBClient(db),
		testplan.WithBundle(bdl),
		testplan.WithTestCase(testCaseSvc),
		testplan.WithTestSet(testSetSvc),
		testplan.WithAutoTest(autotest),
	)

	sonarMetricRule := sonar_metric_rule.New(
		sonar_metric_rule.WithDBClient(db),
		sonar_metric_rule.WithBundle(bdl),
	)

	migrateSvc := migrate.New(migrate.WithDBClient(db))

	ep := endpoints.New(
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithBundle(bdl),
		endpoints.WithDB(db),
		endpoints.WithTestcase(testCaseSvc),
		endpoints.WithTestSet(testSetSvc),
		endpoints.WithSonarMetricRule(sonarMetricRule),
		endpoints.WithTestplan(testPlan),
		endpoints.WithCQ(cq.New(cq.WithBundle(bdl))),
		endpoints.WithAutoTest(autotest),
		endpoints.WithAutoTestV2(autotestV2),
		endpoints.WithSceneSet(sceneset),
		endpoints.WithMigrate(migrateSvc),
	)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())

	loadMetricKeysFromDb(db)

	return server.ListenAndServe()
}

func do(ep *endpoints.Endpoints) error {
	// 注册 webhooks
	if err := ep.RegisterWebhooks(); err != nil {
		return err
	}

	// 根据环境变量填充 maven settings.xml
	renderMvnSettings()

	if err := registerGittarHook(); err != nil {
		logrus.Error(err)
	}

	// TODO: goroutine监听自动创建
	for i := 0; i < conf.ConsumerNum(); i++ {
		go endpoints.StartHookTaskConsumer()
	}

	return nil
}

func registerGittarHook() error {
	req := struct {
		Name       string `json:"name"`
		URL        string `json:"url"`
		PushEvents bool   `json:"push_events"`
	}{
		Name:       "qa",
		URL:        fmt.Sprint("http://", conf.SelfAddr(), "/callback/gittar"),
		PushEvents: true,
	}

	r, err := httpclient.New().
		Post(discover.Gittar()).
		Path("/_system/hooks").
		Header("Content-Type", "application/json").
		JSONBody(&req).Do().DiscardBody()

	if err != nil {
		logrus.Errorf("failed to register gittar, (%+v)", err)
		return err
	}
	if !r.IsOK() {
		return errors.Errorf("failed to register gittar, code: %d", r.StatusCode())
	}

	return nil
}

var configMap = map[string]string{}

// 填充 /root/.m2/settings 文件
func renderMvnSettings() {
	configMap["BP_NEXUS_URL"] = "http://" + conf.NexusAddr()
	configMap["BP_NEXUS_USERNAME"] = conf.NexusUsername()
	configMap["BP_NEXUS_PASSWORD"] = conf.NexusPassword()

	m2File := "/root/.m2/settings.xml"
	bytes, err := ioutil.ReadFile(m2File)
	if err != nil {
		logrus.Errorf("read maven file fail %v", err)
	} else {
		result, _ := renderConfig(string(bytes))
		ioutil.WriteFile(m2File, []byte(result), os.ModePerm)
	}
}

func renderConfig(template string) (string, bool) {
	compile, _ := regexp.Compile("{{.+?}}")
	hasChange := false
	result := compile.ReplaceAllStringFunc(template, func(s string) string {
		key := s[2:(len(s) - 2)]
		value, ok := configMap[key]
		if ok {
			hasChange = true
			return value
		} else {
			return s
		}
	})
	return result, hasChange
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
