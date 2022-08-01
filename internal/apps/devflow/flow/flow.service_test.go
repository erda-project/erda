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

	"github.com/erda-project/erda-infra/providers/i18n"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	flowrulepb "github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
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

type issueForMakeMrDescMock struct {
	IssueMock
}

type transForMakeMrDescMock struct {
}

func (t transForMakeMrDescMock) Get(lang i18n.LanguageCodes, key, def string) string {
	panic("implement me")
}

func (t transForMakeMrDescMock) Text(lang i18n.LanguageCodes, key string) string {
	return "Task"
}

func (t transForMakeMrDescMock) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	panic("implement me")
}

func (i issueForMakeMrDescMock) GetIssue(id int64, identityInfo *commonpb.IdentityInfo) (*issuepb.Issue, error) {
	if id == 0 {
		return nil, fmt.Errorf("error")
	}
	return &issuepb.Issue{Title: "New issue to Erda", Id: 100001}, nil
}

var mRDesc = "Task: #100001 New issue to Erda"

func TestService_makeMrDescByIssueID(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		issueID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "test with error",
			fields: fields{p: &provider{Trans: transForMakeMrDescMock{}, Issue: issueForMakeMrDescMock{}}},
			args: args{
				ctx:     context.Background(),
				issueID: 0,
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "test with no error",
			fields: fields{p: &provider{Trans: transForMakeMrDescMock{}, Issue: issueForMakeMrDescMock{}}},
			args: args{
				ctx:     context.Background(),
				issueID: 1,
			},
			want:    mRDesc,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.makeMrDescByIssueID(tt.args.ctx, tt.args.issueID)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeMrDesc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeMrDesc() got = %v, want %v", got, tt.want)
			}
		})
	}
}
