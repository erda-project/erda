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

package mock

import (
	"context"

	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
)

type CronMock struct {
}

func (c CronMock) CronCreate(ctx context.Context, request *cronpb.CronCreateRequest) (*cronpb.CronCreateResponse, error) {
	panic("implement me")
}

func (c CronMock) CronPaging(ctx context.Context, request *cronpb.CronPagingRequest) (*cronpb.CronPagingResponse, error) {
	panic("implement me")
}

func (c CronMock) CronStart(ctx context.Context, request *cronpb.CronStartRequest) (*cronpb.CronStartResponse, error) {
	panic("implement me")
}

func (c CronMock) CronStop(ctx context.Context, request *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
	panic("implement me")
}

func (c CronMock) CronDelete(ctx context.Context, request *cronpb.CronDeleteRequest) (*cronpb.CronDeleteResponse, error) {
	panic("implement me")
}

func (c CronMock) CronGet(ctx context.Context, request *cronpb.CronGetRequest) (*cronpb.CronGetResponse, error) {
	panic("implement me")
}

func (c CronMock) CronUpdate(ctx context.Context, request *cronpb.CronUpdateRequest) (*cronpb.CronUpdateResponse, error) {
	panic("implement me")
}
