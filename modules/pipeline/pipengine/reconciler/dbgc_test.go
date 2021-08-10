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

package reconciler

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestReconciler_doPipelineDatabaseGC(t *testing.T) {

	var pipelineMaps = map[uint64]spec.Pipeline{
		1: {
			PipelineBase: spec.PipelineBase{ID: 1},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: apistructs.PipelineGC{
					DatabaseGC: apistructs.PipelineDatabaseGC{
						Analyzed: apistructs.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
						Finished: apistructs.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
					},
				},
			}},
		},
		2: {
			PipelineBase: spec.PipelineBase{ID: 2},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: apistructs.PipelineGC{
					DatabaseGC: apistructs.PipelineDatabaseGC{
						Analyzed: apistructs.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
						Finished: apistructs.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
					},
				},
			}},
		},
	}

	var db *dbclient.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "PageListPipelines", func(client *dbclient.Client, req apistructs.PipelinePageListRequest, ops ...dbclient.SessionOption) ([]spec.Pipeline, []uint64, int64, int64, error) {
		assert.True(t, req.PageNum <= 2, "PageNum > 2")
		if req.PageNum == 1 {
			return []spec.Pipeline{pipelineMaps[1], pipelineMaps[0]}, nil, 2, 2, nil
		} else {
			return nil, nil, 0, 0, nil
		}
	})

	var r Reconciler
	r.dbClient = db

	var gcNum = 1
	var addCountNum = func() {
		gcNum++
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(&r), "DoDBGC", func(r *Reconciler, pipelineID uint64, gcOption apistructs.PipelineGCDBOption) error {
		assert.True(t, gcNum < 3, "DoDBGC times >= 3")
		addCountNum()
		return nil
	})

	r.doPipelineDatabaseGC(apistructs.PipelinePageListRequest{
		PageNum:  1,
		PageSize: 10,
	})
}
