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
	"fmt"
	"reflect"
	"testing"

	flowrulepb "github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/devflowrule"
)

type devFlowRuleForGetMock struct {
	devFlowRuleMock
}

var flowRule = &flowrulepb.FlowWithBranchPolicy{
	Flow: &flowrulepb.Flow{
		Name:         "DEV",
		TargetBranch: "feature/*",
		Artifact:     "alpha",
		Environment:  "dev",
	},
	BranchPolicy: &flowrulepb.BranchPolicy{
		Branch:     "feature/*",
		BranchType: "multi_branch",
		Policy: &flowrulepb.Policy{
			SourceBranch:  "develop",
			CurrentBranch: "feature/*",
			TempBranch:    "next/develop",
			TargetBranch: &flowrulepb.TargetBranch{
				MergeRequest: "master",
				CherryPick:   "",
			},
		},
	},
}

func (d devFlowRuleForGetMock) GetFlowByRule(ctx context.Context, request devflowrule.GetFlowByRuleRequest) (*flowrulepb.FlowWithBranchPolicy, error) {
	if request.ProjectID == 0 {
		return nil, fmt.Errorf("error")
	}
	return flowRule, nil
}

func TestService_getFlowRule(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx       context.Context
		projectID uint64
		mrInfo    *apistructs.MergeRequestInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *flowrulepb.FlowWithBranchPolicy
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:       context.Background(),
				projectID: 0,
				mrInfo:    &apistructs.MergeRequestInfo{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:       context.Background(),
				projectID: 1,
				mrInfo:    &apistructs.MergeRequestInfo{},
			},
			want:    flowRule,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.getFlowRule(tt.args.ctx, tt.args.projectID, tt.args.mrInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFlowRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFlowRule() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTempBranchFromFlowRule(t *testing.T) {
	type args struct {
		flowRule *flowrulepb.FlowWithBranchPolicy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty",
			args: args{flowRule: nil},
			want: "",
		},
		{
			name: "test with not empty",
			args: args{flowRule: flowRule},
			want: "next/develop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTempBranchFromFlowRule(tt.args.flowRule); got != tt.want {
				t.Errorf("getTempBranchFromFlowRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFlowNameFromFlowRule(t *testing.T) {
	type args struct {
		flowRule *flowrulepb.FlowWithBranchPolicy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty",
			args: args{flowRule: nil},
			want: "",
		},
		{
			name: "test with not empty",
			args: args{flowRule: flowRule},
			want: "DEV",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFlowNameFromFlowRule(tt.args.flowRule); got != tt.want {
				t.Errorf("getFlowNameFromFlowRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSourceBranchFromFlowRule(t *testing.T) {
	type args struct {
		flowRule *flowrulepb.FlowWithBranchPolicy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty",
			args: args{flowRule: nil},
			want: "",
		},
		{
			name: "test with not empty",
			args: args{flowRule: flowRule},
			want: "develop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSourceBranchFromFlowRule(tt.args.flowRule); got != tt.want {
				t.Errorf("getSourceBranchFromFlowRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
