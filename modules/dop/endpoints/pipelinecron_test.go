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
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert"

	pb1 "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	commonpb "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type TestPipelineCron struct{}

func (t TestPipelineCron) CronCreate(ctx context.Context, request *cronpb.CronCreateRequest) (*cronpb.CronCreateResponse, error) {
	return &cronpb.CronCreateResponse{
		Data: &commonpb.Cron{
			ID: 123,
		},
	}, nil
}

func (t TestPipelineCron) CronPaging(ctx context.Context, request *cronpb.CronPagingRequest) (*cronpb.CronPagingResponse, error) {
	panic("implement me")
}

func (t TestPipelineCron) CronStart(ctx context.Context, request *cronpb.CronStartRequest) (*cronpb.CronStartResponse, error) {
	panic("implement me")
}

func (t TestPipelineCron) CronStop(ctx context.Context, request *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
	panic("implement me")
}

func (t TestPipelineCron) CronDelete(ctx context.Context, request *cronpb.CronDeleteRequest) (*cronpb.CronDeleteResponse, error) {
	panic("implement me")
}

func (t TestPipelineCron) CronGet(ctx context.Context, request *cronpb.CronGetRequest) (*cronpb.CronGetResponse, error) {
	panic("implement me")
}

func (t TestPipelineCron) CronUpdate(ctx context.Context, request *cronpb.CronUpdateRequest) (*cronpb.CronUpdateResponse, error) {
	panic("implement me")
}

func TestEndpoints_pipelineCronCreate(t *testing.T) {
	e := &Endpoints{}
	var createV2 pb1.PipelineCreateRequest
	createV2.PipelineYml = `version: 1.1
cron_compensator:
  enable: true
  latest_first: true
  stop_if_latter_executed: true
stages: []
`
	req, err := http.NewRequest("method", "body", strings.NewReader(jsonparse.JsonOneLine(createV2)))
	assert.Nil(t, err)
	req.Header = map[string][]string{
		"USER-ID": {"2"},
		httputil.InternalHeader: {
			"true",
		},
	}

	var testPipelineCron TestPipelineCron
	e.PipelineCron = testPipelineCron
	got, err := e.pipelineCronCreate(context.Background(), req, nil)
	assert.NoError(t, err)
	resp := got.GetContent().(httpserver.Resp)
	assert.Equal(t, resp.Data, &commonpb.Cron{
		ID: 123,
	})
}
