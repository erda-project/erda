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

package flow

import (
	"context"

	"github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/devflowrule"
)

type devFlowRuleMock struct {
}

func (d devFlowRuleMock) CreateDevFlowRule(ctx context.Context, request *pb.CreateDevFlowRuleRequest) (*pb.CreateDevFlowRuleResponse, error) {
	panic("implement me")
}

func (d devFlowRuleMock) DeleteDevFlowRule(ctx context.Context, request *pb.DeleteDevFlowRuleRequest) (*pb.DeleteDevFlowRuleResponse, error) {
	panic("implement me")
}

func (d devFlowRuleMock) UpdateDevFlowRule(ctx context.Context, request *pb.UpdateDevFlowRuleRequest) (*pb.UpdateDevFlowRuleResponse, error) {
	panic("implement me")
}

func (d devFlowRuleMock) GetDevFlowRulesByProjectID(ctx context.Context, request *pb.GetDevFlowRuleRequest) (*pb.GetDevFlowRuleResponse, error) {
	panic("implement me")
}

func (d devFlowRuleMock) GetFlowByRule(ctx context.Context, request devflowrule.GetFlowByRuleRequest) (*pb.FlowWithBranchPolicy, error) {
	panic("implement me")
}
