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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

func CreateCheckRun(ctx *webcontext.Context) {
	var request apistructs.CheckRun
	err := ctx.BindJSON(&request)
	if err != nil {
		ctx.Abort(err)
		return
	}
	result, err := ctx.Service.CreateOrUpdateCheckRun(ctx.Repository, &request)
	if err != nil {
		ctx.Abort(err)
		return
	}
	ctx.Success(result)
}

func QueryCheckRuns(ctx *webcontext.Context) {
	mrID := ctx.Query("mrId")
	result, err := ctx.Service.QueryCheckRuns(ctx.Repository, mrID)
	if err != nil {
		ctx.Abort(err)
		return
	}
	ctx.Success(result)
}
