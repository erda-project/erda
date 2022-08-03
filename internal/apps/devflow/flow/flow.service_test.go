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

	"bou.ke/monkey"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	flowrulepb "github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
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

func (d devFlowRuleForGetMock) GetDevFlowRulesByProjectID(ctx context.Context, request *flowrulepb.GetDevFlowRuleRequest) (*flowrulepb.GetDevFlowRuleResponse, error) {
	if request.ProjectID == 0 {
		return nil, fmt.Errorf("error")
	}
	return &flowrulepb.GetDevFlowRuleResponse{
		Data: &flowrulepb.DevFlowRule{
			ID: "c1dcf304-0dd6-4e2c-b68a-2005d45e81fd",
			Flows: []*flowrulepb.Flow{
				{
					Name:         "DEV",
					TargetBranch: "feature/*",
					Artifact:     "alpha",
					Environment:  "DEV",
				}, {
					Name:         "TEST",
					TargetBranch: "develop",
					Artifact:     "beta",
					Environment:  "TEST",
				}, {
					Name:         "STAGING",
					TargetBranch: "release/*",
					Artifact:     "rc",
					Environment:  "STAGING",
				}, {
					Name:         "PROD",
					TargetBranch: "master",
					Artifact:     "stable",
					Environment:  "PROD",
				},
			},
			OrgID:       1,
			OrgName:     "erda",
			ProjectID:   5,
			ProjectName: "erda-project",
			BranchPolicies: []*flowrulepb.BranchPolicy{
				{
					Branch:     "master",
					BranchType: "single_branch",
					Policy: &flowrulepb.Policy{
						SourceBranch:  "",
						CurrentBranch: "master",
						TempBranch:    "",
						TargetBranch:  nil,
					},
				}, {
					Branch:     "release/*",
					BranchType: "multi_branch",
					Policy: &flowrulepb.Policy{
						SourceBranch:  "develop",
						CurrentBranch: "release/*",
						TempBranch:    "next/release",
						TargetBranch: &flowrulepb.TargetBranch{
							MergeRequest: "master",
							CherryPick:   "develop",
						},
					},
				}, {
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
				{
					Branch:     "develop",
					BranchType: "single_branch",
					Policy: &flowrulepb.Policy{
						SourceBranch:  "",
						CurrentBranch: "develop",
						TempBranch:    "",
						TargetBranch:  nil,
					},
				},
			},
		},
	}, nil
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

var mRDesc = "Task: #100001 New issue to Erda"

func TestService_makeMrDesc(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx   context.Context
		issue *issuepb.Issue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{p: &provider{Trans: transForMakeMrDescMock{}}},
			args: args{
				ctx:   context.Background(),
				issue: &issuepb.Issue{Title: "New issue to Erda", Id: 100001},
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
			got := s.makeMrDesc(tt.args.ctx, tt.args.issue)
			if got != tt.want {
				t.Errorf("makeMrDesc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeGittarRepoPath(t *testing.T) {
	type args struct {
		app *apistructs.ApplicationDTO
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with nil",
			args: args{
				app: nil,
			},
			want: "",
		}, {
			name: "test with correct",
			args: args{
				app: &apistructs.ApplicationDTO{
					Name:        "erda",
					ProjectName: "erda-project",
				},
			},
			want: "/wb/erda-project/erda",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeGittarRepoPath(tt.args.app); got != tt.want {
				t.Errorf("makeGittarRepoPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isRefPatternMatch(t *testing.T) {
	type args struct {
		currentBranch string
		branchRule    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test with not match",
			args: args{
				currentBranch: "feature/dop",
				branchRule:    "feat/*",
			},
			want: false,
		},
		{
			name: "test with match",
			args: args{
				currentBranch: "feature/dop",
				branchRule:    "feature/*",
			},
			want: true,
		},
		{
			name: "test with match2",
			args: args{
				currentBranch: "feature/dop",
				branchRule:    "feature/*,bugfix/*",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRefPatternMatch(tt.args.currentBranch, tt.args.branchRule); got != tt.want {
				t.Errorf("isRefPatternMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_getFlowRuleNameBranchPolicyMap(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx       context.Context
		projectID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]branchPolicy
		wantErr bool
	}{
		{
			name: "test with err",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:       context.Background(),
				projectID: 0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with correct",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:       context.Background(),
				projectID: 1,
			},
			want: map[string]branchPolicy{
				"DEV": {
					tempBranch:   "next/develop",
					targetBranch: "master",
					sourceBranch: "develop",
				},
				"TEST": {
					tempBranch:   "",
					targetBranch: "",
					sourceBranch: "",
				},
				"STAGING": {
					tempBranch:   "next/release",
					targetBranch: "master",
					sourceBranch: "develop",
				},
				"PROD": {
					tempBranch:   "",
					targetBranch: "",
					sourceBranch: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.getFlowRuleNameBranchPolicyMap(tt.args.ctx, tt.args.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFlowRuleNameBranchPolicyMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFlowRuleNameBranchPolicyMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_findBranchPolicyByName(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx          context.Context
		projectID    uint64
		flowRuleName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *flowrulepb.BranchPolicy
		wantErr bool
	}{
		{
			name: "test with correct",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:          context.Background(),
				projectID:    1,
				flowRuleName: "DEV",
			},
			want: &flowrulepb.BranchPolicy{
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
			wantErr: false,
		},
		{
			name: "test with err",
			fields: fields{
				p: &provider{DevFlowRule: devFlowRuleForGetMock{}},
			},
			args: args{
				ctx:          context.Background(),
				projectID:    1,
				flowRuleName: "DEV2",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.findBranchPolicyByName(tt.args.ctx, tt.args.projectID, tt.args.flowRuleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("findBranchPolicyByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findBranchPolicyByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_isMROpenedOrNotCreated(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMergeRequest",
		func(bdl *bundle.Bundle, appID uint64, userID string, req apistructs.GittarQueryMrRequest) (*apistructs.QueryMergeRequestsData, error) {
			if appID == 0 {
				return nil, fmt.Errorf("fail")
			}
			if appID == 1 {
				return &apistructs.QueryMergeRequestsData{
					List:  nil,
					Total: 0,
				}, nil
			}
			if appID == 2 {
				return &apistructs.QueryMergeRequestsData{
					List: []*apistructs.MergeRequestInfo{
						{
							Id:    1,
							State: "closed",
						},
					},
					Total: 1,
				}, nil
			}
			return &apistructs.QueryMergeRequestsData{
				List: []*apistructs.MergeRequestInfo{
					{
						Id:    2,
						State: "open",
					},
					{
						Id:    3,
						State: "merged",
					},
				},
				Total: 2,
			}, nil
		})
	defer monkey.UnpatchAll()

	type fields struct {
		p *provider
	}
	type args struct {
		ctx           context.Context
		currentBranch string
		targetBranch  string
		appID         uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test with err",
			fields: fields{
				p: &provider{bdl: bdl},
			},
			args: args{
				ctx:           context.Background(),
				currentBranch: "feature/dop",
				targetBranch:  "master",
				appID:         0,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "test with not created",
			fields: fields{
				p: &provider{bdl: bdl},
			},
			args: args{
				ctx:           context.Background(),
				currentBranch: "feature/dop",
				targetBranch:  "master",
				appID:         1,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test with not have open",
			fields: fields{
				p: &provider{bdl: bdl},
			},
			args: args{
				ctx:           context.Background(),
				currentBranch: "feature/dop",
				targetBranch:  "master",
				appID:         2,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "test with have open",
			fields: fields{
				p: &provider{bdl: bdl},
			},
			args: args{
				ctx:           context.Background(),
				currentBranch: "feature/dop",
				targetBranch:  "master",
				appID:         3,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.isMROpenedOrNotCreated(tt.args.ctx, tt.args.currentBranch, tt.args.targetBranch, tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf("isMROpenedOrNotCreated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isMROpenedOrNotCreated() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_getMrInfo(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMergeRequest",
		func(bdl *bundle.Bundle, appID uint64, userID string, req apistructs.GittarQueryMrRequest) (*apistructs.QueryMergeRequestsData, error) {
			if appID == 0 {
				return nil, fmt.Errorf("fail")
			}
			return &apistructs.QueryMergeRequestsData{
				List: []*apistructs.MergeRequestInfo{
					{
						Id:    2,
						State: "open",
					},
					{
						Id:    3,
						State: "merged",
					},
				},
				Total: 2,
			}, nil
		})
	defer monkey.UnpatchAll()

	type fields struct {
		p *provider
	}
	type args struct {
		ctx           context.Context
		appID         uint64
		currentBranch string
		targetBranch  string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantMrInfo *apistructs.MergeRequestInfo
		wantErr    bool
	}{
		{
			name:   "test with err",
			fields: fields{p: &provider{bdl: bdl}},
			args: args{
				ctx:           context.Background(),
				appID:         0,
				currentBranch: "feature/dop",
				targetBranch:  "master",
			},
			wantMrInfo: nil,
			wantErr:    true,
		},
		{
			name:   "test with correct",
			fields: fields{p: &provider{bdl: bdl}},
			args: args{
				ctx:           context.Background(),
				appID:         1,
				currentBranch: "feature/dop",
				targetBranch:  "master",
			},
			wantMrInfo: &apistructs.MergeRequestInfo{
				Id:    2,
				State: "open",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			gotMrInfo, err := s.getMrInfo(tt.args.ctx, tt.args.appID, tt.args.currentBranch, tt.args.targetBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMrInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMrInfo, tt.wantMrInfo) {
				t.Errorf("getMrInfo() gotMrInfo = %v, want %v", gotMrInfo, tt.wantMrInfo)
			}
		})
	}
}

func Test_canJoin(t *testing.T) {
	type args struct {
		commit     *apistructs.Commit
		baseCommit *apistructs.Commit
		tempBranch string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test with nil commit",
			args: args{
				commit:     nil,
				baseCommit: &apistructs.Commit{},
				tempBranch: "next/dev",
			},
			want: false,
		},
		{
			name: "test with empty tempBranch",
			args: args{
				commit:     &apistructs.Commit{},
				baseCommit: &apistructs.Commit{},
				tempBranch: "",
			},
			want: false,
		},
		{
			name: "test with nil baseCommit",
			args: args{
				commit:     &apistructs.Commit{},
				baseCommit: nil,
				tempBranch: "next/dev",
			},
			want: true,
		},
		{
			name: "test with equal commitID",
			args: args{
				commit:     &apistructs.Commit{ID: "1"},
				baseCommit: &apistructs.Commit{ID: "1"},
				tempBranch: "next/dev",
			},
			want: false,
		},
		{
			name: "test with unequal commitID",
			args: args{
				commit:     &apistructs.Commit{ID: "1"},
				baseCommit: &apistructs.Commit{ID: "2"},
				tempBranch: "next/dev",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canJoin(tt.args.commit, tt.args.baseCommit, tt.args.tempBranch); got != tt.want {
				t.Errorf("canJoin() = %v, want %v", got, tt.want)
			}
		})
	}
}

type issueForCreateEventMock struct {
	IssueMock
}

func (i issueForCreateEventMock) GetIssue(id int64, identityInfo *commonpb.IdentityInfo) (*issuepb.Issue, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return &issuepb.Issue{Title: "New issue to Erda", Id: 100001}, nil
}

func TestService_CreateFlowEvent(t *testing.T) {
	var bdl *bundle.Bundle
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(d *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				Name: "project",
			}, nil
		},
	)
	defer p1.Unpatch()

	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateEvent",
		func(d *bundle.Bundle, ev *apistructs.EventCreateRequest) error {
			return nil
		},
	)
	defer p2.Unpatch()

	type args struct {
		req *CreateFlowRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				req: &CreateFlowRequest{
					ProjectID: 1,
					AppID:     2,
					OrgID:     3,
					Data: &pb.FlowEventData{
						IssueID: 1,
					},
				},
			},
		},
	}
	s := &Service{p: &provider{bdl: bdl, Issue: issueForCreateEventMock{}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.CreateFlowEvent(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateFlowEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
