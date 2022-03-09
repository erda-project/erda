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

package dop

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/modules/dop/endpoints"
	"github.com/erda-project/erda/modules/pipeline/providers/cron"
)

func TestCompensatePipelineCms(t *testing.T) {
	ep := endpoints.New()
	p := &provider{}
	var svc *cron.Service
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "CronPaging",
		func(*cron.Service, context.Context, *pb.CronPagingRequest) (*pb.CronPagingResponse, error) {
			return &pb.CronPagingResponse{
				Total: 3,
				Data: []*common.Cron{
					{
						UserID: "1",
						OrgID:  1,
					},
					{
						UserID: "1",
						OrgID:  0,
					},
					{
						UserID: "",
						OrgID:  1,
					},
				},
			}, nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "CronUpdate",
		func(*cron.Service, context.Context, *pb.CronUpdateRequest) (*pb.CronUpdateResponse, error) {
			return nil, nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(ep), "UpdateCmsNsConfigs",
		func(*endpoints.Endpoints, string, uint64) error {
			return nil
		})
	p.PipelineCron = svc
	defer monkey.UnpatchAll()
	err := p.compensatePipelineCms(ep)
	assert.NoError(t, err)
}
