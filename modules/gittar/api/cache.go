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
