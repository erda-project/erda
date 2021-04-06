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
