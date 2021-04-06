package api

import (
	"github.com/erda-project/erda/modules/gittar/migration"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

func MigrationNewAuth(ctx *webcontext.Context) {
	err := migration.NewAuth()
	if err != nil {
		ctx.Abort(err)
		return
	}
	ctx.Success("")
}
