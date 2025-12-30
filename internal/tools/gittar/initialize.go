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

package gittar

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/gittar/api"
	"github.com/erda-project/erda/internal/tools/gittar/auth"
	"github.com/erda-project/erda/internal/tools/gittar/cache"
	"github.com/erda-project/erda/internal/tools/gittar/conf"
	"github.com/erda-project/erda/internal/tools/gittar/helper"
	"github.com/erda-project/erda/internal/tools/gittar/metrics"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/admin_token"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gc"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/internal/tools/gittar/profiling"
	"github.com/erda-project/erda/internal/tools/gittar/rpcmetrics"
	"github.com/erda-project/erda/internal/tools/gittar/uc"
	"github.com/erda-project/erda/internal/tools/gittar/webcontext"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/reconciler/rutil"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Initialize 初始化应用启动服务.
func (p *provider) Initialize(ctx context.Context) error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("DEBUG MODE")
	}

	// Start zombie process reaper (only active if PID 1)
	helper.StartZombieReaper()

	if _, err := os.Stat(conf.RepoRoot()); os.IsNotExist(err) {
		logrus.Infof("repository folder is not exist, will auto create later")
		err := os.MkdirAll(conf.RepoRoot(), 0755)
		if err != nil {
			panic(err)
		}
		logrus.Infof("repository folder created!")
	}

	tok, err := admin_token.InitAdminAuthToken(filepath.Join(conf.RepoRoot(), ".gittar", "auth_token"))
	if err != nil {
		panic(err)
	}
	logrus.Infof("gittar admin auth token: %s", tok)

	p.initGitRPCMetrics()

	gitmodule.Setting.MaxGitDiffLineCharacters = conf.GitMaxDiffLineCharacters()
	gitmodule.Setting.MaxGitDiffSize = conf.GitMaxDiffSize()
	gitmodule.Setting.ContextLineCount = conf.GitDiffContextLines()
	gitmodule.Setting.MaxGitDiffFiles = conf.GitMaxDiffFiles()
	gitmodule.Setting.MaxGitDiffLines = conf.GitMaxDiffLines()

	diceBundle := bundle.New(
		bundle.WithErdaServer(),
	)

	dbClient, err := models.OpenDB()
	if err != nil {
		panic(err)
	}
	uc.InitializeUcClient(p.UserSvc)

	webcontext.WithDB(dbClient)
	webcontext.WithBundle(diceBundle)
	webcontext.WithEtcdClient(p.EtcdClient)
	webcontext.WithTokenService(&p.TokenService)
	webcontext.WithOrgClient(p.Org)
	webcontext.WithI18n(p.I18n)

	e := echo.New()

	svc := models.NewService(dbClient, diceBundle, p.I18n, nil)
	p.initLegacyMetrics(e, svc)

	systemGroup := e.Group("/_system")
	{
		systemGroup.GET("/cache/stats", webcontext.WrapHandler(api.ShowCacheStats))
		systemGroup.GET("/rpc-metrics", webcontext.WrapHandler(api.RPCMetrics))
		systemGroup.POST("/hooks", webcontext.WrapHandler(api.AddSystemHook))
		systemGroup.POST("/repos", webcontext.WrapHandler(api.CreateRepo))
		systemGroup.DELETE("/repos/:id", webcontext.WrapHandler(api.DeleteRepo))
		systemGroup.DELETE("/apps/:id", webcontext.WrapHandler(api.DeleteRepoByApp))
		systemGroup.PUT("/apps/:id", webcontext.WrapHandler(api.UpdateRepoByApp))

		systemGroup.POST("/migration/new_auth", webcontext.WrapHandler(api.MigrationNewAuth))
	}

	apiGroup := e.Group("/_api")
	{
		// implements the health check
		apiGroup.GET("/health", webcontext.WrapHandler(api.Health))
	}

	debugGroup := e.Group("/_debug")
	profiling.WrapGroup(debugGroup)

	gitApiGroup := e.Group("/:org/:repo", webcontext.WrapMiddlewareHandler(auth.Authenticate))
	addApiRoutes(gitApiGroup)

	gitApiGroupNew := e.Group("/app-repo/:appId", webcontext.WrapMiddlewareHandler(auth.AuthenticateByApp))
	addApiRoutes(gitApiGroupNew)

	gitApiGroupV2 := e.Group("/wb/:project/:app", webcontext.WrapMiddlewareHandler(auth.AuthenticateV2))
	addApiRoutes(gitApiGroupV2)

	gitApiGroupV3 := e.Group("/:org/dop/:project/:app", webcontext.WrapMiddlewareHandler(auth.AuthenticateV3))
	addApiRoutes(gitApiGroupV3)

	functionalGroup := e.Group("/api")
	{
		functionalGroup.GET("/merge-requests-count", webcontext.WrapHandler(api.MergeRequestCount))
	}

	logger := middleware.Logger()
	e.Use(logger)

	gitmodule.Setting.RepoStatsCache = cache.NewMysqlCache("repo-stats", dbClient)

	// cron task to git gc all repository
	go gc.ScheduledExecuteClean()

	// start hook task consumer
	models.Init(dbClient)

	// graceful shutdown
	go func() {
		<-ctx.Done()
		logrus.Infof("shutting down http server...")
		if err := e.Shutdown(context.Background()); err != nil {
			logrus.Errorf("failed to shutdown http server: %v", err)
		}
	}()

	return e.Start(":" + conf.ListenPort())
}

func addApiRoutes(g *echo.Group) {
	g.DELETE("", webcontext.WrapHandler(api.DeleteRepo))

	g.GET("", webcontext.WrapHandler(api.GetGoImportMeta))

	// implements the get_text_file function
	g.GET("/HEAD", webcontext.WrapHandler(api.GetRepoHead))
	// implements the get_info_refs function
	g.GET("/info/refs", webcontext.WrapHandler(api.GetRepoInfoRefs))
	// implements the get_loose_object get_pack_file get_text_file function
	g.GET("/objects/:prefix/:suffix", webcontext.WrapHandler(api.GetRepoObjects))
	// implements the service_rpc function
	g.POST("/git-:service", webcontext.WrapHandler(api.ServiceRepoRPC))

	g.GET("/commits/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoCommits))
	g.POST("/commits", webcontext.WrapHandler(api.CreateCommit))

	g.GET("/commit/:sha", webcontext.WrapHandler(api.Commit))
	g.GET("/branches", webcontext.WrapHandler(api.GetRepoBranches))
	g.POST("/branches", webcontext.WrapHandler(api.CreateRepoBranch))
	g.GET("/branches/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoBranchDetail))
	g.DELETE("/branches/*", webcontext.WrapHandler(api.DeleteRepoBranch))
	g.PUT("/branch/default/*", webcontext.WrapHandler(api.SetRepoDefaultBranch))
	g.POST("/locked", webcontext.WrapHandler(api.SetLocked))
	g.GET("/stats/*", webcontext.WrapHandler(api.GetRepoStats))
	g.GET("/stats", webcontext.WrapHandler(api.GetRepoStats))
	g.GET("/tags", webcontext.WrapHandler(api.GetRepoTags))
	g.POST("/tags", webcontext.WrapHandler(api.CreateRepoTag))
	g.DELETE("/tags/*", webcontext.WrapHandler(api.DeleteRepoTag))
	g.GET("/tree/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoTree))
	g.GET("/tree-search", webcontext.WrapHandlerWithRepoCheck(api.SearchRepoTree))
	g.GET("/blob/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoBlob))
	g.GET("/blob-range/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoBlobRange))
	g.GET("/raw/*", webcontext.WrapHandlerWithRepoCheck(api.GetRepoRaw))
	g.GET("/blame/*", webcontext.WrapHandlerWithRepoCheck(api.BlameFile))

	g.GET("/compare/*", webcontext.WrapHandlerWithRepoCheck(api.Compare))

	g.GET("/diff-file", webcontext.WrapHandlerWithRepoCheck(api.DiffFile))

	// merge request
	g.GET("/merge-stats", webcontext.WrapHandler(api.CheckMergeStatus))
	g.GET("/merge-templates", webcontext.WrapHandler(api.GetMergeTemplates))
	g.GET("/merge-requests/:id", webcontext.WrapHandler(api.GetMergeRequestDetail))
	g.GET("/merge-requests", webcontext.WrapHandler(api.GetMergeRequests))
	g.GET("/merge-request-stats", webcontext.WrapHandler(api.GetMergeRequestsStats))
	g.POST("/merge-requests", webcontext.WrapHandler(api.CreateMergeRequest))
	g.POST("/merge-requests/:id/edit", webcontext.WrapHandler(api.UpdateMergeRequest))
	g.POST("/merge-requests/:id/merge", webcontext.WrapHandler(api.Merge))
	g.POST("/merge-requests/:id/close", webcontext.WrapHandler(api.CloseMR))
	g.POST("/merge-requests/:id/reopen", webcontext.WrapHandler(api.ReopenMR))
	g.GET("/merge-requests/:id/notes", webcontext.WrapHandler(api.QueryNotes))
	g.POST("/merge-requests/:id/notes", webcontext.WrapHandler(api.CreateNotes))
	g.POST("/merge-requests/:id/operation-temp-branch", webcontext.WrapHandler(api.OperationTempBranch))
	g.POST("/check-runs", webcontext.WrapHandler(api.CreateCheckRun))
	g.GET("/check-runs", webcontext.WrapHandler(api.QueryCheckRuns))
	g.POST("/merge-with-branch", webcontext.WrapHandler(api.MergeWithBranch))
	g.GET("/merge-base", webcontext.WrapHandler(api.GetMergeBase))

	// web hooks
	g.GET("/hooks", webcontext.WrapHandler(api.GetHooks))
	g.POST("/hooks", webcontext.WrapHandler(api.AddHook))
	g.GET("/hooks/:id", webcontext.WrapHandler(api.GetHookDetail))
	g.PUT("/hooks/:id", webcontext.WrapHandler(api.UpdateHook))
	g.DELETE("/hooks/:id", webcontext.WrapHandler(api.DeleteHook))

	// files manage
	g.POST("/backup/*", webcontext.WrapHandler(api.Backup))
	g.GET("/backup-list", webcontext.WrapHandler(api.BackupList))
	g.DELETE("/backup/*", webcontext.WrapHandler(api.DeleteBackup))
	g.GET("/archive/*", webcontext.WrapHandlerWithRepoCheck(api.GetArchive))

}

func (p *provider) initGitRPCMetrics() {
	err := rpcmetrics.Init(rpcmetrics.Config{
		Enabled:       conf.GitRPCMetricsEnabled(),
		Path:          conf.GitRPCMetricsPath(),
		BufferSize:    conf.GitRPCMetricsBuffer(),
		FlushInterval: 1 * time.Second,
	})
	if err != nil {
		logrus.Errorf("failed to init git rpc metrics writer: %v", err)
		return
	}

	// When metrics enabled and writing to a directory, start an internal cleanup loop.
	// It removes daily files older than retention days every 6 hours.
	retentionDays := conf.GitRPCMetricsRetentionDays()
	if conf.GitRPCMetricsEnabled() && retentionDays > 0 && rpcmetrics.IsDirPath(conf.GitRPCMetricsPath()) {
		rpcmetrics.StartCleanupLoop(rpcmetrics.CleanupConfig{
			Dir:      conf.GitRPCMetricsPath(),
			KeepDays: retentionDays,
			Interval: 6 * time.Hour,
		})
	}
}

func (p *provider) initLegacyMetrics(e *echo.Echo, svc *models.Service) {
	collector := metrics.NewCollector(svc)
	go func() {
		<-time.NewTimer(time.Duration(rand.Intn(5)) * time.Minute).C
		rutil.ContinueWorking(context.Background(), p.Log, func(ctx context.Context) rutil.WaitDuration {
			logrus.Infof("start refresh personal contributors")
			if err := collector.RefreshPersonalContributions(); err != nil {
				logrus.Errorf("failed to refresh personal contributors, err: %v", err)
			}
			return rutil.ContinueWorkingWithDefaultInterval
		}, rutil.WithContinueWorkingDefaultRetryInterval(conf.RefreshPersonalContributorDuration()))
	}()

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)
	e.GET("/metrics", func(ctx echo.Context) error {
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	})
	e.POST("/personal-contribution", webcontext.WrapHandler(func(c *webcontext.Context) {
		var req apistructs.GittarListRepoRequest
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatus(400, fmt.Errorf("request body parse failed, err: %v", err))
			return
		}
		contributors, err := collector.IterateRepos(req)
		if err != nil {
			c.AbortWithStatus(400, fmt.Errorf("failed to list personal contributors, err: %v", err))
			return
		}
		c.Success(contributors)
	}))
}
