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
