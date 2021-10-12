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

package dop

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	infrahttpserver "github.com/erda-project/erda-infra/providers/httpserver"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/endpoints"
	"github.com/erda-project/erda/modules/dop/event"
	"github.com/erda-project/erda/modules/dop/services/apidocsvc"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/appcertificate"
	"github.com/erda-project/erda/modules/dop/services/assetsvc"
	"github.com/erda-project/erda/modules/dop/services/autotest"
	atv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"github.com/erda-project/erda/modules/dop/services/branchrule"
	"github.com/erda-project/erda/modules/dop/services/cdp"
	"github.com/erda-project/erda/modules/dop/services/certificate"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	"github.com/erda-project/erda/modules/dop/services/comment"
	"github.com/erda-project/erda/modules/dop/services/cq"
	"github.com/erda-project/erda/modules/dop/services/environment"
	"github.com/erda-project/erda/modules/dop/services/filetree"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/dop/services/issuefilterbm"
	"github.com/erda-project/erda/modules/dop/services/issuepanel"
	"github.com/erda-project/erda/modules/dop/services/issueproperty"
	"github.com/erda-project/erda/modules/dop/services/issuerelated"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/modules/dop/services/issuestream"
	"github.com/erda-project/erda/modules/dop/services/iteration"
	"github.com/erda-project/erda/modules/dop/services/libreference"
	"github.com/erda-project/erda/modules/dop/services/migrate"
	"github.com/erda-project/erda/modules/dop/services/monitor"
	"github.com/erda-project/erda/modules/dop/services/namespace"
	"github.com/erda-project/erda/modules/dop/services/nexussvc"
	"github.com/erda-project/erda/modules/dop/services/org"
	"github.com/erda-project/erda/modules/dop/services/permission"
	"github.com/erda-project/erda/modules/dop/services/pipeline"
	"github.com/erda-project/erda/modules/dop/services/projectpipelinefiletree"
	"github.com/erda-project/erda/modules/dop/services/publisher"
	"github.com/erda-project/erda/modules/dop/services/sceneset"
	"github.com/erda-project/erda/modules/dop/services/sonar_metric_rule"
	"github.com/erda-project/erda/modules/dop/services/testcase"
	"github.com/erda-project/erda/modules/dop/services/testplan"
	"github.com/erda-project/erda/modules/dop/services/testset"
	"github.com/erda-project/erda/modules/dop/services/ticket"
	"github.com/erda-project/erda/modules/dop/services/workbench"
	"github.com/erda-project/erda/modules/dop/utils"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

const EtcdPipelineCmsCompensate = "dop/pipelineCms/compensate"

// Initialize 初始化应用启动服务.
func (p *provider) Initialize(ctx servicehub.Context) error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("set log level: %s", logrus.DebugLevel)
	}

	// TODO invoke self use service
	//_ = os.Setenv("QA_ADDR", discover.QA())

	db := &dao.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: p.DB,
		},
	}
	dbclient.Set(db.DBEngine)
	ep, err := p.initEndpoints(db)
	if err != nil {
		return err
	}

	//定时上报issue
	go monitor.TimedTaskMetricsAddAndRepairBug(ep.DBClient(), bdl.Bdl)
	go monitor.TimedTaskMetricsIssue(ep.DBClient(), ep.UCClient(), bdl.Bdl)

	registerWebHook(bdl.Bdl)

	go endpoints.SetProjectStatsCache()

	// 注册 hook
	if err := ep.RegisterEvents(); err != nil {
		return err
	}

	p.Protocol.WithContextValue(types.IssueStateService, ep.IssueStateService())
	p.Protocol.WithContextValue(types.IssueFilterBmService, issuefilterbm.New(
		issuefilterbm.WithDBClient(db),
	))
	p.Protocol.WithContextValue(types.CodeCoverageService, ep.CodeCoverageService())
	p.Protocol.WithContextValue(types.IssueService, ep.IssueService())

	// This server will never be started. Only the routes and locale loader are used by new http server
	server := httpserver.New(":0")
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("cmdb"))
	server.WithLocaleLoader(bdl.Bdl.GetLocaleLoader())
	server.Router().PathPrefix("/api/apim/metrics").Handler(endpoints.InternalReverseHandler(endpoints.ProxyMetrics))
	ctx.Service("http-server").(infrahttpserver.Router).Any("/**", server.Router())

	loadMetricKeysFromDb(db)

	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())

	interval := time.Duration(conf.TestFileIntervalSec())
	purgeCycle := conf.TestFileRecordPurgeCycleDay()
	if err := ep.TestCaseService().BatchClearProcessingRecords(); err != nil {
		logrus.Error(err)
		return err
	}
	// Scheduled polling export task
	go func() {
		ticker := time.NewTicker(time.Second * interval)
		for {
			select {
			case <-ticker.C:
				exportTestFileTask(ep)
			case <-ep.ExportChannel:
				exportTestFileTask(ep)
			}
		}
	}()

	// Scheduled polling import task
	go func() {
		ticker := time.NewTicker(time.Second * interval)
		for {
			select {
			case <-ticker.C:
				importTestFileTask(ep)
			case <-ep.ImportChannel:
				importTestFileTask(ep)
			}
		}
	}()

	// Scheduled polling copy task
	go func() {
		ticker := time.NewTicker(time.Second * interval)
		for {
			select {
			case <-ticker.C:
				copyTestFileTask(ep)
			case <-ep.CopyChannel:
				copyTestFileTask(ep)
			}
		}
	}()

	// Daily clear test file records
	go func() {
		day := time.NewTicker(time.Hour * 24 * time.Duration(purgeCycle))
		for {
			select {
			case <-day.C:
				if err := ep.TestCaseService().DeleteRecordApiFilesByTime(time.Now().AddDate(0, 0, -purgeCycle)); err != nil {
					logrus.Error(err)
				}
			}
		}
	}()

	// compensate pipeline cms according to pipeline cron which enable is true
	go func() {
		// add etcd lock to ensure that it is executed only once
		resp, err := p.EtcdClient.Get(context.Background(), EtcdPipelineCmsCompensate)
		if err != nil {
			logrus.Error(err)
			return
		}
		if len(resp.Kvs) == 0 {
			logrus.Infof("start compensate pipelineCms")
			if err = compensatePipelineCms(ep); err != nil {
				logrus.Error(err)
			}
			_, err = p.EtcdClient.Put(context.Background(), EtcdPipelineCmsCompensate, "true")
			if err != nil {
				logrus.Error(err)
			}
		}

	}()

	// daily issue expiry status update cron job
	go func() {
		cron := cron.New()
		err := cron.AddFunc(conf.UpdateIssueExpiryStatusCron(), func() {
			start := time.Now()
			if err := ep.DBClient().BatchUpdateIssueExpiryStatus(apistructs.StateBelongs); err != nil {
				logrus.Errorf("daily issue expiry status batch update err: %v", err)
				return
			}
			logrus.Infof("daily issue expiry status batch update takes %v", time.Since(start))
		})
		if err != nil {
			panic(err)
		}
		cron.Start()
	}()

	return nil
}

func (p *provider) initEndpoints(db *dao.DBClient) (*endpoints.Endpoints, error) {
	var (
		etcdStore *etcd.Store
		ossClient *oss.Client
		store     jsonstore.JsonStore
	)

	etcdStore, err := etcd.New()
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

	store, err = jsonstore.New()
	if err != nil {
		return nil, err
	}

	c := cdp.New(cdp.WithBundle(bdl.Bdl))

	// init event
	e := event.New(event.WithBundle(bdl.Bdl))

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

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

	autotest := autotest.New(autotest.WithDBClient(db),
		autotest.WithBundle(bdl.Bdl),
		autotest.WithPipelineCms(p.PipelineCms),
	)

	sceneset := sceneset.New(
		sceneset.WithDBClient(db),
		sceneset.WithBundle(bdl.Bdl),
	)

	autotestV2 := atv2.New(
		atv2.WithDBClient(db),
		atv2.WithBundle(bdl.Bdl),
		atv2.WithSceneSet(sceneset),
		atv2.WithAutotestSvc(autotest),
		atv2.WithPipelineCms(p.PipelineCms),
	)

	autotestV2.UpdateFileRecord = testCaseSvc.UpdateFileRecord
	autotestV2.CreateFileRecord = testCaseSvc.CreateFileRecord

	p.TestPlanSvc.WithAutoTestSvc(autotestV2)

	sceneset.GetScenes = autotestV2.ListAutotestScene
	sceneset.CopyScene = autotestV2.CopyAutotestScene

	sonarMetricRule := sonar_metric_rule.New(
		sonar_metric_rule.WithDBClient(db),
		sonar_metric_rule.WithBundle(bdl.Bdl),
	)

	migrateSvc := migrate.New(migrate.WithDBClient(db))

	// 初始化UC Client
	uc := ucauth.NewUCClient(discover.UC(), conf.UCClientID(), conf.UCClientSecret())
	if conf.OryEnabled() {
		uc = ucauth.NewUCClient(conf.OryKratosPrivateAddr(), conf.OryCompatibleClientID(), conf.OryCompatibleClientSecret())
		uc.SetDBClient(db.DB)
	}

	// init ticket service
	t := ticket.New(ticket.WithDBClient(db),
		ticket.WithBundle(bdl.Bdl),
	)

	// init comment service
	com := comment.New(
		comment.WithDBClient(db),
		comment.WithBundle(bdl.Bdl),
	)

	ns := namespace.New(
		namespace.WithDBClient(db),
		namespace.WithBundle(bdl.Bdl),
	)

	branchRule := branchrule.New(
		branchrule.WithDBClient(db),
		branchrule.WithBundle(bdl.Bdl),
	)
	gittarFileTreeSvc := filetree.New(filetree.WithBundle(bdl.Bdl), filetree.WithBranchRule(branchRule))

	// 查询
	pFileTree := projectpipelinefiletree.New(
		projectpipelinefiletree.WithBundle(bdl.Bdl),
		projectpipelinefiletree.WithFileTreeSvc(gittarFileTreeSvc),
		projectpipelinefiletree.WithAutoTestSvc(autotest),
	)

	// init permission
	perm := permission.New(permission.WithBundle(bdl.Bdl), permission.WithBranchRule(branchRule))

	filetreeSvc := apidocsvc.New(apidocsvc.WithBranchRuleSvc(branchRule))

	env := environment.New(
		environment.WithDBClient(db),
		environment.WithBundle(bdl.Bdl),
	)

	issueRelated := issuerelated.New(
		issuerelated.WithDBClient(db),
		issuerelated.WithBundle(bdl.Bdl),
	)

	issueStream := issuestream.New(
		issuestream.WithDBClient(db),
		issuestream.WithBundle(bdl.Bdl),
	)

	issueproperty := issueproperty.New(
		issueproperty.WithDBClient(db),
		issueproperty.WithBundle(bdl.Bdl),
	)

	issue := issue.New(
		issue.WithDBClient(db),
		issue.WithBundle(bdl.Bdl),
		issue.WithIssueStream(issueStream),
		issue.WithUCClient(uc),
		issue.WithTranslator(p.IssueTan),
		issue.WithIssueRelated(issueRelated),
	)

	issueState := issuestate.New(
		issuestate.WithDBClient(db),
		issuestate.WithBundle(bdl.Bdl),
	)

	issuePanel := issuepanel.New(
		issuepanel.WithDBClient(db),
		issuepanel.WithBundle(bdl.Bdl),
		issuepanel.WithIssue(issue),
	)

	itr := iteration.New(
		iteration.WithDBClient(db),
		iteration.WithIssue(issue),
	)

	testPlan := testplan.New(
		testplan.WithDBClient(db),
		testplan.WithBundle(bdl.Bdl),
		testplan.WithTestCase(testCaseSvc),
		testplan.WithTestSet(testSetSvc),
		testplan.WithAutoTest(autotest),
		testplan.WithIssue(issue),
		testplan.WithIssueState(issueState),
	)

	workBench := workbench.New(
		workbench.WithBundle(bdl.Bdl),
		workbench.WithIssue(issue),
	)

	rsaCrypt := encryption.NewRSAScrypt(encryption.RSASecret{
		PublicKey:          conf.Base64EncodedRsaPublicKey(),
		PublicKeyDataType:  encryption.Base64,
		PrivateKey:         conf.Base64EncodedRsaPrivateKey(),
		PrivateKeyDataType: encryption.Base64,
		PrivateKeyType:     encryption.PKCS1,
	})

	// init nexus service
	nexusSvc := nexussvc.New(
		nexussvc.WithDBClient(db),
		nexussvc.WithBundle(bdl.Bdl),
		nexussvc.WithRsaCrypt(rsaCrypt),
		nexussvc.WithPipelineCms(p.PipelineCms),
	)

	// init publisher service
	pub := publisher.New(
		publisher.WithDBClient(db),
		publisher.WithUCClient(uc),
		publisher.WithBundle(bdl.Bdl),
		publisher.WithNexusSvc(nexusSvc),
	)

	// init certificate service
	cer := certificate.New(
		certificate.WithDBClient(db),
		certificate.WithBundle(bdl.Bdl),
	)

	// init appcertificate service
	appCer := appcertificate.New(
		appcertificate.WithDBClient(db),
		appcertificate.WithBundle(bdl.Bdl),
		appcertificate.WithCertificate(cer),
		appcertificate.WithPipelineCms(p.PipelineCms),
	)

	libReference := libreference.New(
		libreference.WithDBClient(db),
		libreference.WithBundle(bdl.Bdl),
	)

	// init org service
	o := org.New(
		org.WithDBClient(db),
		org.WithUCClient(uc),
		org.WithBundle(bdl.Bdl),
		org.WithPublisher(pub),
		org.WithNexusSvc(nexusSvc),
	)

	codeCvc := code_coverage.New(
		code_coverage.WithDBClient(db),
		code_coverage.WithBundle(bdl.Bdl),
		code_coverage.WithEnvConfig(env),
	)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithBundle(bdl.Bdl),
		endpoints.WithPipeline(pipeline.New(
			pipeline.WithBundle(bdl.Bdl),
			pipeline.WithBranchRuleSvc(branchRule),
			pipeline.WithPublisherSvc(pub),
			pipeline.WithPipelineCms(p.PipelineCms),
			pipeline.WithPipelineDefinitionServices(p.PipelineDs),
		)),
		endpoints.WithPipelineCms(p.PipelineCms),
		endpoints.WithEvent(e),
		endpoints.WithCDP(c),
		endpoints.WithPermission(perm),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithGittarFileTree(gittarFileTreeSvc),
		endpoints.WithProjectPipelineFileTree(pFileTree),

		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithAssetSvc(assetsvc.New(assetsvc.WithBranchRuleSvc(branchRule))),
		endpoints.WithFileTreeSvc(filetreeSvc),

		endpoints.WithDB(db),
		endpoints.WithTestcase(testCaseSvc),
		endpoints.WithTestSet(testSetSvc),
		endpoints.WithSonarMetricRule(sonarMetricRule),
		endpoints.WithTestplan(testPlan),
		endpoints.WithWorkbench(workBench),
		endpoints.WithCQ(cq.New(cq.WithBundle(bdl.Bdl), cq.WithBranchRule(branchRule))),
		endpoints.WithAutoTest(autotest),
		endpoints.WithAutoTestV2(autotestV2),
		endpoints.WithSceneSet(sceneset),
		endpoints.WithMigrate(migrateSvc),

		endpoints.WithJSONStore(store),
		endpoints.WithEtcdStore(etcdStore),
		endpoints.WithOSSClient(ossClient),
		endpoints.WithTicket(t),
		endpoints.WithComment(com),
		endpoints.WithBranchRule(branchRule),
		endpoints.WithNamespace(ns),
		endpoints.WithEnvConfig(env),
		endpoints.WithIssue(issue),
		endpoints.WithIssueRelated(issueRelated),
		endpoints.WithIssueStream(issueStream),
		endpoints.WithIssueProperty(issueproperty),
		endpoints.WithIssueState(issueState),
		endpoints.WithIssuePanel(issuePanel),
		endpoints.WithIteration(itr),
		endpoints.WithPublisher(pub),
		endpoints.WithCertificate(cer),
		endpoints.WithAppCertificate(appCer),
		endpoints.WithLibReference(libReference),
		endpoints.WithOrg(o),
		endpoints.WithCodeCoverageExecRecord(codeCvc),
	)

	ep.ImportChannel = make(chan uint64)
	ep.ExportChannel = make(chan uint64)
	ep.CopyChannel = make(chan uint64)
	return ep, nil
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

func registerWebHook(bdl *bundle.Bundle) {
	// 注册审批流状态变更监听
	ev := apistructs.CreateHookRequest{
		Name:   "dop_approve_status_changed",
		Events: []string{bundle.ApprovalStatusChangedEvent},
		URL:    strutil.Concat("http://", discover.DOP(), "/api/approvals/actions/watch-status"),
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

	ev = apistructs.CreateHookRequest{
		Name:   "pipeline_yml_update",
		Events: []string{bundle.GitPushEvent},
		URL:    strutil.Concat("http://", discover.DOP(), "/api/cicd-crons/actions/hook-for-update"),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register pipeline yml event, %v", err)
	}

	ev = apistructs.CreateHookRequest{
		Name:   "pipeline_definition_update",
		Events: []string{bundle.GitPushEvent},
		URL:    strutil.Concat("http://", discover.DOP(), "/api/cicd-pipelines/actions/hook-for-definition-update"),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register pipeline yml event, %v", err)
	}
}

func exportTestFileTask(ep *endpoints.Endpoints) {
	svc := ep.TestCaseService()
	ok, record, err := svc.GetFirstFileReady(apistructs.FileActionTypeExport, apistructs.FileSpaceActionTypeExport)
	if err != nil {
		logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		return
	}
	if !ok {
		return
	}
	switch record.Type {
	case apistructs.FileActionTypeExport:
		svc.ExportFile(record)
	case apistructs.FileSpaceActionTypeExport:
		at2Svc := ep.AutotestV2Service()
		at2Svc.ExportFile(record)
	default:

	}
}

func importTestFileTask(ep *endpoints.Endpoints) {
	svc := ep.TestCaseService()
	ok, record, err := svc.GetFirstFileReady(apistructs.FileActionTypeImport, apistructs.FileSpaceActionTypeImport)
	if err != nil {
		logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		return
	}
	if !ok {
		return
	}
	switch record.Type {
	case apistructs.FileActionTypeImport:
		svc.ImportFile(record)
	case apistructs.FileSpaceActionTypeImport:
		at2Svc := ep.AutotestV2Service()
		at2Svc.ImportFile(record)
	default:

	}
}

func copyTestFileTask(ep *endpoints.Endpoints) {
	ok, record, err := ep.TestCaseService().GetFirstFileReady(apistructs.FileActionTypeCopy)
	if err != nil {
		logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		return
	}
	if !ok {
		return
	}
	ep.TestSetService().CopyTestSet(record)
}

// compensatePipelineCms compensate pipeline cms according to pipeline cron which enable is true
// it will be deprecated in the later version
func compensatePipelineCms(ep *endpoints.Endpoints) error {
	enable := true
	// get total
	cron, err := bdl.Bdl.PageListPipelineCrons(apistructs.PipelineCronPagingRequest{
		AllSources: false,
		Sources:    []apistructs.PipelineSource{apistructs.PipelineSourceDice},
		YmlNames:   nil,
		PageSize:   1,
		PageNo:     1,
		Enable:     &enable,
	})
	if err != nil {
		logrus.Errorf("failed to PageListPipelineCrons, err: %s", err.Error())
		return err
	}
	total := cron.Total
	pageSize := 1000
	crons := make([]*apistructs.PipelineCronDTO, 0, total)
	for i := 0; i < int(total)/pageSize+1; i++ {
		cron, err = bdl.Bdl.PageListPipelineCrons(apistructs.PipelineCronPagingRequest{
			AllSources: true,
			Sources:    nil,
			YmlNames:   nil,
			PageSize:   pageSize,
			PageNo:     i + 1,
			Enable:     &enable,
		})
		if err != nil {
			logrus.Errorf("failed to PageListPipelineCrons, err: %s", err.Error())
			return err
		}
		crons = append(crons, cron.Data...)
	}

	// userOrgMap judge the user ns is compensated or not in the org
	// key: userID-orgID, value: struct{}
	userOrgMap := make(map[string]struct{})
	for _, v := range crons {
		if v.Enable != nil && *v.Enable &&
			v.UserID != "" && v.OrgID != 0 {
			ns := utils.MakeUserOrgPipelineCmsNs(v.UserID, v.OrgID)
			if !strutil.InSlice(ns, v.ConfigManageNamespaces) {
				err := bdl.Bdl.UpdatePipelineCron(apistructs.PipelineCronUpdateRequest{
					ID:                     v.ID,
					PipelineYml:            v.PipelineYml,
					CronExpr:               v.CronExpr,
					ConfigManageNamespaces: []string{utils.MakeUserOrgPipelineCmsNs(v.UserID, v.OrgID)},
				})
				if err != nil {
					logrus.Errorf("failed to UpdatePipelineCron, err: %s", err.Error())
				}
			}
			if _, ok := userOrgMap[fmt.Sprintf("%s-%d", v.UserID, v.OrgID)]; !ok {
				userOrgMap[fmt.Sprintf("%s-%d", v.UserID, v.OrgID)] = struct{}{}
				// the member may not exist
				err = ep.UpdateCmsNsConfigs(v.UserID, v.OrgID)
				if err != nil {
					logrus.Errorf("failed to UpdateCmsNsConfigs, err: %s", err.Error())
				}
			}
		}
	}
	return nil
}
