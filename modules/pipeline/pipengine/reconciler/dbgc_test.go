// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler

import (
	"context"
	"fmt"
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

func TestMakeDBGCKey(t *testing.T) {
	pipelineID := uint64(123)
	gcKey := makeDBGCKey(pipelineID)
	assert.Equal(t, "/devops/pipeline/dbgc/pipeline/123", gcKey)
}

func TestMakeDBGCDLockKey(t *testing.T) {
	pipelineID := uint64(123)
	lockKey := makeDBGCDLockKey(pipelineID)
	assert.Equal(t, "/devops/pipeline/dbgc/dlock/123", lockKey)
}

func TestGetPipelineIDFromDBGCWatchedKey(t *testing.T) {
	key := "/devops/pipeline/dbgc/pipeline/123"
	pipelineID, err := getPipelineIDFromDBGCWatchedKey(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(123), pipelineID)

	key = "/devops/pipeline/dbgc/pipeline/xxx"
	pipelineID, err = getPipelineIDFromDBGCWatchedKey(key)
	assert.Equal(t, false, err == nil)
	assert.Equal(t, uint64(0), pipelineID)
}

func TestPipelineDatabaseGC(t *testing.T) {
	var r Reconciler
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(&r), "PipelineDatabaseGC", func(r *Reconciler, ctx context.Context) {
		return
	})
	defer pm.Unpatch()
	t.Run("PipelineDatabaseGC", func(t *testing.T) {
		r.PipelineDatabaseGC(context.Background())
	})
}

func TestReconciler_doPipelineDatabaseGC1(t *testing.T) {
	t.Run("test", func(t *testing.T) {

		var dbClient *dbclient.Client
		var r Reconciler
		r.dbClient = dbClient
		patch := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "PageListPipelines", func(db *dbclient.Client, req apistructs.PipelinePageListRequest, ops ...dbclient.SessionOption) ([]spec.Pipeline, []uint64, int64, int64, error) {
			switch req.PageNum {
			case 1:
				return nil, nil, 0, 0, fmt.Errorf("error")
			case 2:
				return []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{},
						PipelineExtra: spec.PipelineExtra{
							PipelineID: 1,
						},
					},
				}, nil, 0, 0, nil
			default:
				return []spec.Pipeline{}, nil, 0, 0, nil
			}
		})
		defer patch.Unpatch()

		patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(&r), "DoDBGC", func(r *Reconciler, pipelineID uint64, gcOption apistructs.PipelineGCDBOption) error {
			assert.Equal(t, pipelineID, uint64(1))
			return fmt.Errorf("error")
		})
		defer patch1.Unpatch()

		r.doPipelineDatabaseGC(apistructs.PipelinePageListRequest{PageNum: 1})
	})
}
