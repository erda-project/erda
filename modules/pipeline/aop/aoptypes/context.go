package aoptypes

import (
	"context"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type TuneContext struct {
	context.Context
	SDK SDK
}

type SDK struct {
	Bundle   *bundle.Bundle
	DBClient *dbclient.Client
	Report   *reportsvc.ReportSvc

	TuneType    TuneType
	TuneTrigger TuneTrigger

	Pipeline spec.Pipeline
	Task     spec.PipelineTask
}

func (sdk SDK) Clone() SDK {
	return SDK{
		Bundle:   sdk.Bundle,
		DBClient: sdk.DBClient,
		Report:   sdk.Report,
	}
}

const (
	CtxKeyTasks = iota
)

func (ctx *TuneContext) PutKV(k, v interface{}) {
	ctx.Context = context.WithValue(ctx.Context, k, v)
}

func (ctx *TuneContext) TryGet(k interface{}) (interface{}, bool) {
	v := ctx.Context.Value(k)
	if v == nil {
		return nil, false
	}
	return v, true
}
