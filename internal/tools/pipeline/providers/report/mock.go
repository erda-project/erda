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

package report

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/pipeline/report/pb"
)

type MockReport struct{}

func (m *MockReport) QueryPipelineReportSet(context.Context, *pb.PipelineReportSetQueryRequest) (*pb.PipelineReportSetQueryResponse, error) {
	return nil, nil
}
func (m *MockReport) PagingPipelineReportSet(context.Context, *pb.PipelineReportSetPagingRequest) (*pb.PipelineReportSetPagingResponse, error) {
	return nil, nil
}
func (m *MockReport) Create(req *pb.PipelineReportCreateRequest) (*pb.PipelineReport, error) {
	return nil, nil
}
func (m *MockReport) MakePBMeta(meta interface{}) (*structpb.Struct, error) {
	var reqMeta map[string]interface{}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &reqMeta); err != nil {
		return nil, err
	}
	pbMeta, err := structpb.NewStruct(reqMeta)
	if err != nil {
		return nil, err
	}
	return pbMeta, nil
}
