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

package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/services/pipeline"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/mock"
)

func TestEndpoints_projectPipelineDetail(t *testing.T) {
	type fields struct {
		result       pipelinepb.PipelineDetailDTO
		request      pipelinepb.PipelineDetailRequest
		assertUserID uint64
	}
	type args struct {
		r    *http.Request
		vars map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    httpserver.Responser
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				request: pipelinepb.PipelineDetailRequest{
					PipelineID:               1,
					SimplePipelineBaseResult: false,
				},
				assertUserID: 1,
				result: pipelinepb.PipelineDetailDTO{
					ID:            1,
					ApplicationID: 1,
					Branch:        "master",
				},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{
						RawQuery: "pipelineID=1&simplePipelineBaseResult=false",
					},
					Header: http.Header{
						"USER-ID":         []string{"1"},
						"Internal-Client": []string{"pipeline"},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoints{}

			queryStringDecoder := schema.NewDecoder()
			queryStringDecoder.IgnoreUnknownKeys(true)
			e.queryStringDecoder = queryStringDecoder

			var pipelineSvc = &mock.MockPipelineServiceServer{}
			monkey.PatchInstanceMethod(reflect.TypeOf(pipelineSvc), "PipelineDetail", func(_ *mock.MockPipelineServiceServer, ctx context.Context, req *pipelinepb.PipelineDetailRequest) (*pipelinepb.PipelineDetailResponse, error) {
				assert.Equal(t, req.PipelineID, tt.fields.request.PipelineID)
				assert.Equal(t, req.SimplePipelineBaseResult, tt.fields.request.SimplePipelineBaseResult)

				return &pipelinepb.PipelineDetailResponse{Data: &tt.fields.result}, nil
			})
			e.PipelineSvc = pipelineSvc

			got, err := e.projectPipelineDetail(context.Background(), tt.args.r, tt.args.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("pipelineDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, got.GetContent().(httpserver.Resp).Data.(*pipelinepb.PipelineDetailDTO).ID, tt.fields.result.ID)
			assert.Equal(t, got.GetContent().(httpserver.Resp).Data.(*pipelinepb.PipelineDetailDTO).Branch, tt.fields.result.Branch)
			assert.Equal(t, got.GetContent().(httpserver.Resp).Data.(*pipelinepb.PipelineDetailDTO).ApplicationID, tt.fields.result.ApplicationID)
		})
	}
}

func TestEndpoints_projectPipelineCreate(t *testing.T) {
	type arg struct {
		req pipelinepb.PipelineCreateRequestV2
	}
	testCases := []struct {
		name string
		arg  arg
	}{
		{
			name: "project pipeline",
			arg: arg{
				req: pipelinepb.PipelineCreateRequestV2{
					Labels: map[string]string{
						apistructs.LabelOrgID:     "1",
						apistructs.LabelOrgName:   "erda",
						apistructs.LabelProjectID: "1",
					},
					PipelineYmlName: "project-pipeline",
					PipelineYml: `version: "1.1"
name: ""
stages:
  - stage:
      - custom-script:
          alias: custom-script
          version: "1.0"
          image: custom-script-action:latest
          commands:
            - sleep 10
          resources:
            cpu: 0.1
            mem: 256`,
				},
			},
		},
	}
	pipelineSvc := &pipeline.Pipeline{}
	monkey.PatchInstanceMethod(reflect.TypeOf(pipelineSvc), "CreatePipelineV2", func(_ *pipeline.Pipeline, reqPipeline *pipelinepb.PipelineCreateRequestV2) (*pb.PipelineDTO, error) {
		return &pb.PipelineDTO{
			ID: 1,
		}, nil
	})
	e := Endpoints{
		pipeline: pipelineSvc,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dat, err := json.Marshal(tc.arg.req)
			if err != nil {
				t.Error(err)
			}
			reader := bytes.NewReader([]byte(dat))
			body := io.NopCloser(reader)
			r := &http.Request{
				Header: http.Header{
					"Internal-Client": []string{"pipeline"},
				},
				Body: body,
			}
			_, err = e.projectPipelineCreate(context.Background(), r, map[string]string{})
			assert.NoError(t, err)
		})
	}
}
