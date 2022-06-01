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
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	api2 "github.com/erda-project/erda/modules/tools/gittar/api"
	"github.com/erda-project/erda/modules/tools/gittar/auth"
	"github.com/erda-project/erda/modules/tools/gittar/cache"
	"github.com/erda-project/erda/modules/tools/gittar/conf"
	models2 "github.com/erda-project/erda/modules/tools/gittar/models"
	"github.com/erda-project/erda/modules/tools/gittar/pkg/gc"
	"github.com/erda-project/erda/modules/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/tools/gittar/profiling"
	"github.com/erda-project/erda/modules/tools/gittar/uc"
	"github.com/erda-project/erda/modules/tools/gittar/webcontext"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/ucauth"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func (p *provider) Initialize() error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("DEBUG MODE")
	}

	if _, err := os.Stat(conf.RepoRoot()); os.IsNotExist(err) {
		logrus.Infof("repository folder is not exist, will auto create later")
		err := os.MkdirAll(conf.RepoRoot(), 0755)
		if err != nil {
			panic(err)
		}
		logrus.Infof("repository folder created!")
	}

	gitmodule.Setting.MaxGitDiffLineCharacters = conf.GitMaxDiffLineCharacters()
	gitmodule.Setting.MaxGitDiffSize = conf.GitMaxDiffSize()
	gitmodule.Setting.ContextLineCount = conf.GitDiffContextLines()
	gitmodule.Setting.MaxGitDiffFiles = conf.GitMaxDiffFiles()
	gitmodule.Setting.MaxGitDiffLines = conf.GitMaxDiffLines()

	ucUserAuth := ucauth.NewUCUserAuth("", discover.UC(), "", conf.UCClientID(), conf.UCClientSecret())
	if conf.OryEnabled() {
		ucUserAuth.ClientID = conf.OryCompatibleClientID()
		ucUserAuth.UCHost = conf.OryKratosAddr()
	}
	diceBundle := bundle.New(
		bundle.WithCoreServices(),
	)

	dbClient, err := models2.OpenDB()
	if err != nil {
		panic(err)
	}
	uc.InitializeUcClient(dbClient.DBEngine.DB)

	webcontext.WithDB(dbClient)
	webcontext.WithBundle(diceBundle)
	webcontext.WithUCAuth(ucUserAuth)
	webcontext.WithEtcdClient(p.EtcdClient)
	webcontext.WithTokenService(&p.TokenService)

	e := echo.New()
	systemGroup := e.Group("/_system")
	{
		systemGroup.GET("/cache/stats", webcontext.WrapHandler(api2.ShowCacheStats))
		systemGroup.POST("/hooks", webcontext.WrapHandler(api2.AddSystemHook))
		systemGroup.POST("/repos", webcontext.WrapHandler(api2.CreateRepo))
		systemGroup.DELETE("/repos/:id", webcontext.WrapHandler(api2.DeleteRepo))
		systemGroup.DELETE("/apps/:id", webcontext.WrapHandler(api2.DeleteRepoByApp))
		systemGroup.PUT("/apps/:id", webcontext.WrapHandler(api2.UpdateRepoByApp))

		systemGroup.POST("/migration/new_auth", webcontext.WrapHandler(api2.MigrationNewAuth))
	}

	apiGroup := e.Group("/_api")
	{
		// implements the health check
		apiGroup.GET("/health", webcontext.WrapHandler(api2.Health))
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
		functionalGroup.GET("/merge-requests-count", webcontext.WrapHandler(api2.MergeRequestCount))
	}

	logger := middleware.Logger()
	e.Use(logger)

	// requestHistogram := promxp.RegisterHistogram(
	// 	"request_duration",
	// 	"gittar API请求耗时",
	// 	map[string]string{}, // labels
	// 	[]float64{0.001, 0.01, 0.1, 1, 3, 5, 10},
	// )
	// e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return func(c echo.Context) error {
	// 		start := time.Now()
	// 		if err := next(c); err != nil {
	// 			return err
	// 		}
	// 		requestHistogram.Observe(time.Since(start).Seconds())
	// 		return nil
	// 	}
	// })
	// monitor
	// e.GET("/metrics", echo.WrapHandler(promxp.Handler("gittar")))

	gitmodule.Setting.RepoStatsCache = cache.NewMysqlCache("repo-stats", dbClient)

	// cron task to git gc all repository
	go gc.ScheduledExecuteClean()

	// start hook task consumer
	models2.Init(dbClient)

	return e.Start(":" + conf.ListenPort())
}

func addApiRoutes(g *echo.Group) {
	g.DELETE("", webcontext.WrapHandler(api2.DeleteRepo))

	g.GET("", webcontext.WrapHandler(api2.GetGoImportMeta))

	// implements the get_text_file function
	g.GET("/HEAD", webcontext.WrapHandler(api2.GetRepoHead))
	// implements the get_info_refs function
	g.GET("/info/refs", webcontext.WrapHandler(api2.GetRepoInfoRefs))
	// implements the get_loose_object get_pack_file get_text_file function
	g.GET("/objects/:prefix/:suffix", webcontext.WrapHandler(api2.GetRepoObjects))
	// implements the service_rpc function
	g.POST("/git-:service", webcontext.WrapHandler(api2.ServiceRepoRPC))

	g.GET("/commits/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoCommits))
	g.POST("/commits", webcontext.WrapHandler(api2.CreateCommit))

	g.GET("/commit/:sha", webcontext.WrapHandler(api2.Commit))
	g.GET("/branches", webcontext.WrapHandler(api2.GetRepoBranches))
	g.POST("/branches", webcontext.WrapHandler(api2.CreateRepoBranch))
	g.GET("/branches/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoBranchDetail))
	g.DELETE("/branches/*", webcontext.WrapHandler(api2.DeleteRepoBranch))
	g.PUT("/branch/default/*", webcontext.WrapHandler(api2.SetRepoDefaultBranch))
	g.POST("/locked", webcontext.WrapHandler(api2.SetLocked))
	g.GET("/stats/*", webcontext.WrapHandler(api2.GetRepoStats))
	g.GET("/stats", webcontext.WrapHandler(api2.GetRepoStats))
	g.GET("/tags", webcontext.WrapHandler(api2.GetRepoTags))
	g.POST("/tags", webcontext.WrapHandler(api2.CreateRepoTag))
	g.DELETE("/tags/*", webcontext.WrapHandler(api2.DeleteRepoTag))
	g.GET("/tree/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoTree))
	g.GET("/tree-search", webcontext.WrapHandlerWithRepoCheck(api2.SearchRepoTree))
	g.GET("/blob/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoBlob))
	g.GET("/blob-range/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoBlobRange))
	g.GET("/raw/*", webcontext.WrapHandlerWithRepoCheck(api2.GetRepoRaw))
	g.GET("/blame/*", webcontext.WrapHandlerWithRepoCheck(api2.BlameFile))

	g.GET("/compare/*", webcontext.WrapHandlerWithRepoCheck(api2.Compare))

	g.GET("/diff-file", webcontext.WrapHandlerWithRepoCheck(api2.DiffFile))

	//merge request
	g.GET("/merge-stats", webcontext.WrapHandler(api2.CheckMergeStatus))
	g.GET("/merge-templates", webcontext.WrapHandler(api2.GetMergeTemplates))
	g.GET("/merge-requests/:id", webcontext.WrapHandler(api2.GetMergeRequestDetail))
	g.GET("/merge-requests", webcontext.WrapHandler(api2.GetMergeRequests))
	g.GET("/merge-request-stats", webcontext.WrapHandler(api2.GetMergeRequestsStats))
	g.POST("/merge-requests", webcontext.WrapHandler(api2.CreateMergeRequest))
	g.POST("/merge-requests/:id/edit", webcontext.WrapHandler(api2.UpdateMergeRequest))
	g.POST("/merge-requests/:id/merge", webcontext.WrapHandler(api2.Merge))
	g.POST("/merge-requests/:id/close", webcontext.WrapHandler(api2.CloseMR))
	g.POST("/merge-requests/:id/reopen", webcontext.WrapHandler(api2.ReopenMR))
	g.GET("/merge-requests/:id/notes", webcontext.WrapHandler(api2.QueryNotes))
	g.POST("/merge-requests/:id/notes", webcontext.WrapHandler(api2.CreateNotes))
	g.POST("/check-runs", webcontext.WrapHandler(api2.CreateCheckRun))
	g.GET("/check-runs", webcontext.WrapHandler(api2.QueryCheckRuns))

	// web hooks
	g.GET("/hooks", webcontext.WrapHandler(api2.GetHooks))
	g.POST("/hooks", webcontext.WrapHandler(api2.AddHook))
	g.GET("/hooks/:id", webcontext.WrapHandler(api2.GetHookDetail))
	g.PUT("/hooks/:id", webcontext.WrapHandler(api2.UpdateHook))
	g.DELETE("/hooks/:id", webcontext.WrapHandler(api2.DeleteHook))

	// files manage
	g.POST("/backup/*", webcontext.WrapHandler(api2.Backup))
	g.GET("/backup-list", webcontext.WrapHandler(api2.BackupList))
	g.DELETE("/backup/*", webcontext.WrapHandler(api2.DeleteBackup))
	g.GET("/archive/*", webcontext.WrapHandlerWithRepoCheck(api2.GetArchive))

}
