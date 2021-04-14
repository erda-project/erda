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

package gittar

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/gittar/api"
	"github.com/erda-project/erda/modules/gittar/auth"
	"github.com/erda-project/erda/modules/gittar/cache"
	"github.com/erda-project/erda/modules/gittar/conf"
	"github.com/erda-project/erda/modules/gittar/models"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/gittar/profiling"
	"github.com/erda-project/erda/modules/gittar/webcontext"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/ucauth"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
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
	diceBundle := bundle.New(
		bundle.WithCMDB(),
		bundle.WithEventBox(),
	)

	dbClient, err := models.OpenDB()
	if err != nil {
		panic(err)
	}
	webcontext.WithDB(dbClient)
	webcontext.WithBundle(diceBundle)
	webcontext.WithUCAuth(ucUserAuth)

	e := echo.New()
	systemGroup := e.Group("/_system")
	{
		systemGroup.GET("/cache/stats", webcontext.WrapHandler(api.ShowCacheStats))
		systemGroup.POST("/hooks", webcontext.WrapHandler(api.AddSystemHook))
		systemGroup.POST("/repos", webcontext.WrapHandler(api.CreateRepo))
		systemGroup.DELETE("/repos/:id", webcontext.WrapHandler(api.DeleteRepo))
		systemGroup.DELETE("/apps/:id", webcontext.WrapHandler(api.DeleteRepoByApp))
		systemGroup.PUT("/apps/:id", webcontext.WrapHandler(api.UpdateRepoByApp))

		systemGroup.POST("/migration/new_auth", webcontext.WrapHandler(api.MigrationNewAuth))
	}

	debugGroup := e.Group("/_debug")
	profiling.WrapGroup(debugGroup)

	gitApiGroup := e.Group("/:org/:repo", webcontext.WrapMiddlewareHandler(auth.Authenticate))
	addApiRoutes(gitApiGroup)

	gitApiGroupNew := e.Group("/app-repo/:appId", webcontext.WrapMiddlewareHandler(auth.AuthenticateByApp))
	addApiRoutes(gitApiGroupNew)

	gitApiGroupV2 := e.Group("/wb/:project/:app", webcontext.WrapMiddlewareHandler(auth.AuthenticateV2))
	addApiRoutes(gitApiGroupV2)

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

	//merge request
	g.GET("/merge-stats", webcontext.WrapHandler(api.CheckMergeStatus))
	g.GET("/merge-templates", webcontext.WrapHandler(api.GetMergeTemplates))
	g.GET("/merge-requests/:id", webcontext.WrapHandler(api.GetMergeRequestDetail))
	g.GET("/merge-requests", webcontext.WrapHandler(api.GetMergeRequests))
	g.POST("/merge-requests", webcontext.WrapHandler(api.CreateMergeRequest))
	g.POST("/merge-requests/:id/edit", webcontext.WrapHandler(api.UpdateMergeRequest))
	g.POST("/merge-requests/:id/merge", webcontext.WrapHandler(api.Merge))
	g.POST("/merge-requests/:id/close", webcontext.WrapHandler(api.CloseMR))
	g.POST("/merge-requests/:id/reopen", webcontext.WrapHandler(api.ReopenMR))
	g.GET("/merge-requests/:id/notes", webcontext.WrapHandler(api.QueryNotes))
	g.POST("/merge-requests/:id/notes", webcontext.WrapHandler(api.CreateNotes))
	g.POST("/check-runs", webcontext.WrapHandler(api.CreateCheckRun))
	g.GET("/check-runs", webcontext.WrapHandler(api.QueryCheckRuns))

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
