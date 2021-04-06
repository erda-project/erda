package api

import (
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

func ShowCacheStats(context *webcontext.Context) {

	context.Success(Map{
		"commit":     gitmodule.Setting.CommitCache.Status(),
		"pathCommit": gitmodule.Setting.PathCommitCache.Status(),
		"repoStats":  gitmodule.Setting.RepoStatsCache.Status(),
		"objectSize": gitmodule.Setting.ObjectSizeCache.Status(),
	})

}
