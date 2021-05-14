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

package aop

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func init() {
	bdl := bundle.New(bundle.WithAllAvailableClients())
	dbClient, _ := dbclient.New()
	report := reportsvc.New(reportsvc.WithDBClient(dbClient))

	Initialize(bdl, dbClient, report)
}

func TestHandlePipeline(t *testing.T) {
	ctx := NewContextForTask(spec.PipelineTask{ID: 1}, spec.Pipeline{}, aoptypes.TuneTriggerTaskBeforeWait)
	err := Handle(ctx)
	assert.NoError(t, err)

	// pipeline end
	ctx = NewContextForPipeline(spec.Pipeline{
		PipelineBase: spec.PipelineBase{ID: 1},
	}, aoptypes.TuneTriggerPipelineAfterExec)
	err = Handle(ctx)
	assert.NoError(t, err)
}

func TestHandleTask(t *testing.T) {
	// task end
	ctx := NewContextForTask(
		spec.PipelineTask{ID: 1, Status: apistructs.PipelineStatusSuccess},
		spec.Pipeline{
			PipelineBase: spec.PipelineBase{ID: 1},
		}, aoptypes.TuneTriggerTaskAfterExec,
	)
	err := Handle(ctx)
	assert.NoError(t, err)
}
