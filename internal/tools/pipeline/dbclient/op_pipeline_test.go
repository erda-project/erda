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

package dbclient

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"xorm.io/xorm"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

func TestNotFoundBaseErrorIs(t *testing.T) {
	e := fmt.Errorf("not found base")
	assert.True(t, errors.Is(NotFoundBaseError, e))

	ne := fmt.Errorf("other error")
	assert.False(t, errors.Is(NotFoundBaseError, ne))
}

func TestPageListPipelines(t *testing.T) {
	sourceLabels := []interface{}{"s1", "s2"}
	labels, _ := structpb.NewValue(sourceLabels)
	tests := []struct {
		name            string
		labels          map[string]*structpb.Value
		wantPipelineIDS []uint64
	}{
		{
			name:            "no labels",
			labels:          map[string]*structpb.Value{},
			wantPipelineIDS: []uint64{1, 2, 3},
		},
		{
			name: "with labels",
			labels: map[string]*structpb.Value{
				"source": labels,
			},
			wantPipelineIDS: []uint64{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbSession := &xorm.Session{}
			session := &Session{Session: dbSession}
			client := &Client{}
			pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(client), "NewSession", func(_ *Client, ops ...SessionOption) *Session {
				return session
			})
			defer pm1.Unpatch()
			pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "Where", func(_ *xorm.Session, query interface{}, args ...interface{}) *xorm.Session {
				return dbSession
			})
			defer pm2.Unpatch()
			pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "Cols", func(_ *xorm.Session, columns ...string) *xorm.Session {
				return dbSession
			})
			defer pm3.Unpatch()
			pm4 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "Table", func(_ *xorm.Session, tableNameOrBean interface{}) *xorm.Session {
				return dbSession
			})
			defer pm4.Unpatch()
			pm5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "Desc", func(_ *xorm.Session, colNames ...string) *xorm.Session {
				return dbSession
			})
			defer pm5.Unpatch()
			pm6 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "FindAndCount", func(_ *xorm.Session, rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
				basePipelines, ok := rowsSlicePtr.(*[]spec.PipelineBase)
				if ok {
					for _, pipelineID := range tt.wantPipelineIDS {
						*basePipelines = append(*basePipelines, spec.PipelineBase{ID: pipelineID})
					}
				}
				total := int64(len(tt.wantPipelineIDS))
				return total, nil
			})
			defer pm6.Unpatch()
			pm7 := monkey.PatchInstanceMethod(reflect.TypeOf(client), "ListPipelinesByIDs", func(_ *Client, pipelineIDs []uint64, needQueryDefinition bool, ops ...SessionOption) ([]spec.Pipeline, error) {
				return []spec.Pipeline{
					{PipelineBase: spec.PipelineBase{ID: 1}},
					{PipelineBase: spec.PipelineBase{ID: 2}},
					{PipelineBase: spec.PipelineBase{ID: 3}},
				}, nil
			})
			defer pm7.Unpatch()
			pm8 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "SQL", func(_ *xorm.Session, query interface{}, args ...interface{}) *xorm.Session {
				return dbSession
			})
			defer pm8.Unpatch()
			pm9 := monkey.PatchInstanceMethod(reflect.TypeOf(dbSession), "Find", func(_ *xorm.Session, rowsSlicePtr interface{}, condiBean ...interface{}) error {
				innerPipelineIDs, ok := rowsSlicePtr.(*[]uint64)
				if ok {
					for _, pipelineID := range tt.wantPipelineIDS {
						*innerPipelineIDs = append(*innerPipelineIDs, pipelineID)
					}
				}
				return nil
			})
			defer pm9.Unpatch()
			pagingRes, err := client.PageListPipelines(&pipelinepb.PipelinePagingRequest{
				AnyMatchLabelsJSON: tt.labels,
				PageNo:             1,
				PageSize:           3,
				AllSources:         true,
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPipelineIDS, pagingRes.PagingPipelineIDs)
		})
	}
}

func TestGetMinPipelineID(t *testing.T) {
	res := &PageListPipelinesResult{
		PagingPipelineIDs: []uint64{11918341, 11918330, 11918310, 11918308, 11918303, 11918291, 11918283, 11917574},
	}
	minID := res.GetMinPipelineID()
	assert.Equal(t, uint64(11917574), minID)
}
