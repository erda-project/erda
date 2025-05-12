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

package dbgc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc/db"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc/dbgcconfig"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

func TestReconciler_doPipelineDatabaseGC(t *testing.T) {
	var pipelines = []spec.Pipeline{
		{
			PipelineBase: spec.PipelineBase{ID: 1, Status: apistructs.PipelineStatusAnalyzed},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: basepb.PipelineGC{
					DatabaseGC: &basepb.PipelineDatabaseGC{
						Analyzed: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
						Finished: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
					},
				},
			}},
		},
		{
			PipelineBase: spec.PipelineBase{ID: 1, Status: apistructs.PipelineStatusAnalyzed},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: basepb.PipelineGC{},
			}},
		},
		{
			PipelineBase: spec.PipelineBase{ID: 2, Status: apistructs.PipelineStatusAnalyzed},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: basepb.PipelineGC{
					DatabaseGC: &basepb.PipelineDatabaseGC{
						Analyzed: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
						Finished: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
					},
				},
			}},
		},
		{
			PipelineBase: spec.PipelineBase{ID: 3, Status: apistructs.PipelineStatusRunning},
			PipelineExtra: spec.PipelineExtra{Extra: spec.PipelineExtraInfo{
				GC: basepb.PipelineGC{
					DatabaseGC: &basepb.PipelineDatabaseGC{
						Analyzed: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
						Finished: &basepb.PipelineDBGCItem{
							NeedArchive: &[]bool{true}[0],
							TTLSecond:   &[]uint64{100}[0],
						},
					},
				},
			}},
		},
	}

	DB := &dbclient.Client{}

	pm := monkey.PatchInstanceMethod(reflect.TypeOf(DB), "PageListPipelines", func(client *dbclient.Client, req *pipelinepb.PipelinePagingRequest, ops ...dbclient.SessionOption) (*dbclient.PageListPipelinesResult, error) {
		assert.True(t, req.PageNum <= 3, "PageNum > 3")
		if len(pipelines) == 0 {
			return &dbclient.PageListPipelinesResult{}, nil
		}
		res := &dbclient.PageListPipelinesResult{
			Pipelines:         pipelines[req.PageSize*(req.PageNum-1) : req.PageSize*req.PageNum],
			PagingPipelineIDs: nil,
			Total:             2,
			CurrentPageSize:   2,
		}
		pipelines = pipelines[req.PageSize*req.PageNum:]
		return res, nil
	})
	defer pm.Unpatch()

	var r provider
	r.dbClient = &db.Client{Client: *DB}

	var gcNum = 1
	var addCountNum = func() {
		gcNum++
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(&r), "DoDBGC", func(r *provider, pipeline *spec.Pipeline, gcOption apistructs.PipelineGCDBOption) error {
		assert.True(t, gcNum < 4, "DoDBGC times >= 4")
		addCountNum()
		return nil
	})
	pm1 := monkey.Patch(time.Sleep, func(d time.Duration) {
		return
	})
	defer pm1.Unpatch()

	r.doPipelineDatabaseGC(context.Background(), &pipelinepb.PipelinePagingRequest{
		PageNum:  1,
		PageSize: 1,
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
	var r provider
	DB := &dbclient.Client{}

	pm := monkey.PatchInstanceMethod(reflect.TypeOf(DB), "PageListPipelines", func(client *dbclient.Client, req *pipelinepb.PipelinePagingRequest, ops ...dbclient.SessionOption) (*dbclient.PageListPipelinesResult, error) {
		return &dbclient.PageListPipelinesResult{Pipelines: nil}, nil
	})
	defer pm.Unpatch()

	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(r.dbClient), "DeletePipelineArchives", func(client *db.Client, req db.ArchiveDeleteRequest, ops ...dbclient.SessionOption) error {
		return nil
	})
	defer pm1.Unpatch()
	r.dbClient = &db.Client{Client: *DB}
	r.Cfg = &dbgcconfig.Config{
		PipelineDBGCDuration: 3 * time.Second,
	}
	r.Log = logrusx.New()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	t.Run("PipelineDatabaseGC", func(t *testing.T) {
		r.PipelineDatabaseGC(ctx)
	})
}

func TestReconciler_doPipelineDatabaseGC1(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		DB := &dbclient.Client{}
		var r provider
		r.dbClient = &db.Client{Client: *DB}
		patch := monkey.PatchInstanceMethod(reflect.TypeOf(DB), "PageListPipelines", func(client *dbclient.Client, req *pipelinepb.PipelinePagingRequest, ops ...dbclient.SessionOption) (*dbclient.PageListPipelinesResult, error) {
			switch req.PageNum {
			case 1:
				return &dbclient.PageListPipelinesResult{
					Pipelines:         nil,
					PagingPipelineIDs: nil,
					Total:             0,
					CurrentPageSize:   0,
				}, nil
			case 2:
				return &dbclient.PageListPipelinesResult{
					Pipelines: []spec.Pipeline{
						{
							PipelineBase: spec.PipelineBase{},
							PipelineExtra: spec.PipelineExtra{
								PipelineID: 1,
							},
						}},
					PagingPipelineIDs: nil,
					Total:             1,
					CurrentPageSize:   1,
				}, nil
			default:
				return &dbclient.PageListPipelinesResult{
					Pipelines:         []spec.Pipeline{},
					PagingPipelineIDs: nil,
					Total:             0,
					CurrentPageSize:   0,
				}, nil
			}
		})
		defer patch.Unpatch()

		patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(&r), "DoDBGC", func(r *provider, pipeline *spec.Pipeline, gcOption apistructs.PipelineGCDBOption) error {
			assert.Equal(t, pipeline.PipelineID, uint64(1))
			return fmt.Errorf("error")
		})
		defer patch1.Unpatch()

		r.doPipelineDatabaseGC(context.Background(), &pipelinepb.PipelinePagingRequest{PageNum: 1})
	})
}

func TestDoDBGC(t *testing.T) {
	p := provider{}
	pipeline := &spec.Pipeline{}
	needArchive := true
	assert.Panics(t, func() {
		p.DoDBGC(pipeline, apistructs.PipelineGCDBOption{NeedArchive: needArchive})
	})
}
