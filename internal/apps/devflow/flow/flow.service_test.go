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
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	flowrulepb "github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/devflow/flow/db"
	"github.com/erda-project/erda/internal/apps/dop/providers/devflowrule"
	"github.com/erda-project/erda/internal/apps/dop/services/permission"
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
			got, err := s.IsMROpenedOrNotCreated(tt.args.ctx, tt.args.currentBranch, tt.args.targetBranch, tt.args.appID)
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

func TestService_idemCreateDevFlow(t *testing.T) {
	var dbClient *db.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetDevFlowByUnique", func(dbClient *db.Client, appID, issueID uint64, branch string) (f *db.DevFlow, err error) {
		if appID == 0 || issueID == 0 || branch == "" {
			return nil, gorm.ErrInvalidData
		}
		if appID == 1 && issueID == 1 && branch == "feature/erda" {
			return &db.DevFlow{
				Model: db.Model{
					ID: fields.UUID{
						String: "157c8320-3755-402b-81ab-01d8bdd99512",
						Valid:  false,
					},
				},
			}, nil
		}
		return nil, gorm.ErrRecordNotFound
	})
	defer monkey.UnpatchAll()

	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "CreateDevFlow", func(dbClient *db.Client, f *db.DevFlow) (err error) {
		f.ID = fields.UUID{
			String: "1d2c6da1-f633-4ff2-8bba-0cdc7043664e",
			Valid:  false,
		}
		return nil
	})

	type field struct {
		p *provider
	}
	type args struct {
		f *db.DevFlow
	}
	tests := []struct {
		name    string
		fields  field
		args    args
		want    *db.DevFlow
		wantErr bool
	}{
		{
			name: "test with err",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				f: &db.DevFlow{
					Model: db.Model{},
					Scope: db.Scope{
						AppID: 0,
					},
					Branch:  "",
					IssueID: 0,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with already exist",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				f: &db.DevFlow{
					Model: db.Model{},
					Scope: db.Scope{
						AppID: 1,
					},
					Branch:  "feature/erda",
					IssueID: 1,
				},
			},
			want: &db.DevFlow{
				Model: db.Model{
					ID: fields.UUID{
						String: "157c8320-3755-402b-81ab-01d8bdd99512",
						Valid:  false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test with not exist",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				f: &db.DevFlow{
					Model: db.Model{},
					Scope: db.Scope{
						AppID: 1,
					},
					Branch:  "feature/erda2",
					IssueID: 1,
				},
			},
			want: &db.DevFlow{
				Model: db.Model{
					ID: fields.UUID{
						String: "1d2c6da1-f633-4ff2-8bba-0cdc7043664e",
						Valid:  false,
					},
				},
				Scope: db.Scope{
					AppID: 1,
				},
				Branch:  "feature/erda2",
				IssueID: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.idemCreateDevFlow(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("idemCreateDevFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("idemCreateDevFlow() got = %v, want %v", got, tt.want)
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

func TestService_listDevFlowByReq(t *testing.T) {
	var dbClient *db.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "ListDevFlowByIssueID", func(dbClient *db.Client, issueID uint64) (f []db.DevFlow, err error) {
		return []db.DevFlow{{
			Model: db.Model{
				ID: fields.UUID{
					String: "157c8320-3755-402b-81ab-01d8bdd99512",
					Valid:  false,
				},
			},
		}}, nil
	})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "ListDevFlowByAppIDAndBranch", func(dbClient *db.Client, appID uint64, branch string) (f []db.DevFlow, err error) {
		if appID == 0 || branch == "" {
			return nil, fmt.Errorf("fail")
		}
		return []db.DevFlow{{
			Model: db.Model{
				ID: fields.UUID{
					String: "1d2c6da1-f633-4ff2-8bba-0cdc7043664e",
					Valid:  false,
				},
			},
		}}, nil
	})

	type field struct {
		p *provider
	}
	type args struct {
		req *pb.GetDevFlowInfoRequest
	}
	tests := []struct {
		name         string
		fields       field
		args         args
		wantDevFlows []db.DevFlow
		wantErr      bool
	}{
		{
			name: "test with zero issueID with error",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				req: &pb.GetDevFlowInfoRequest{
					IssueID: 0,
					AppID:   0,
					Branch:  "",
				},
			},
			wantDevFlows: nil,
			wantErr:      true,
		},
		{
			name: "test with zero issueID",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				req: &pb.GetDevFlowInfoRequest{
					IssueID: 0,
					AppID:   1,
					Branch:  "feature/erda",
				},
			},
			wantDevFlows: []db.DevFlow{{
				Model: db.Model{
					ID: fields.UUID{
						String: "1d2c6da1-f633-4ff2-8bba-0cdc7043664e",
						Valid:  false,
					},
				},
			}},
			wantErr: false,
		},
		{
			name: "test with issueID",
			fields: field{
				p: &provider{dbClient: dbClient},
			},
			args: args{
				req: &pb.GetDevFlowInfoRequest{
					IssueID: 1,
					AppID:   0,
					Branch:  "",
				},
			},
			wantDevFlows: []db.DevFlow{{
				Model: db.Model{
					ID: fields.UUID{
						String: "157c8320-3755-402b-81ab-01d8bdd99512",
						Valid:  false,
					},
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			gotDevFlows, err := s.listDevFlowByReq(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listDevFlowByReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDevFlows, tt.wantDevFlows) {
				t.Errorf("listDevFlowByReq() gotDevFlows = %v, want %v", gotDevFlows, tt.wantDevFlows)
			}
		})
	}
}

func TestService_getAppInIssuePermissionMap(t *testing.T) {
	var perm *permission.Permission

	monkey.PatchInstanceMethod(reflect.TypeOf(perm), "CheckAppAction", func(perm *permission.Permission, identityInfo apistructs.IdentityInfo, appID uint64, action string) error {
		if appID == 1 {
			return nil
		}
		return fmt.Errorf("fail")
	})
	defer monkey.UnpatchAll()

	type fields struct {
		p *provider
	}
	type args struct {
		ctx      context.Context
		devFlows []db.DevFlow
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[uint64]bool
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				p: &provider{
					devFlowService: &Service{permission: perm},
				},
			},
			args: args{
				ctx: context.Background(),
				devFlows: []db.DevFlow{
					{
						Scope: db.Scope{
							AppID: 1,
						},
					},
					{
						Scope: db.Scope{
							AppID: 2,
						},
					},
				},
			},
			want:    map[uint64]bool{1: true, 2: false},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.getAppInIssuePermissionMap(tt.args.ctx, tt.args.devFlows)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAppInIssuePermissionMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAppInIssuePermissionMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_getAppTempBranchCommitAndChangeBranchListMap(t *testing.T) {
	var dbClient *db.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "ListDevFlowByFlowRuleNameAndAppIDs", func(dbClient *db.Client, ruleName string, appIDs ...uint64) (f []db.DevFlow, err error) {
		if ruleName == "" {
			return nil, fmt.Errorf("fail")
		}
		return []db.DevFlow{
			{
				Model: db.Model{
					ID: fields.UUID{
						String: "157c8320-3755-402b-81ab-01d8bdd99512",
						Valid:  false,
					},
				},
				Scope: db.Scope{
					AppID:   1,
					AppName: "erda",
				},
				Branch:           "feature/dop",
				IsJoinTempBranch: true,
			},
			{
				Model: db.Model{
					ID: fields.UUID{
						String: "157c8320-3755-402b-81ab-01d8bdd99666",
						Valid:  false,
					},
				},
				Scope: db.Scope{
					AppID:   1,
					AppName: "erda",
				},
				Branch:           "feature/dop-test",
				IsJoinTempBranch: false,
			},
		}, nil
	})
	defer monkey.UnpatchAll()

	var service *Service
	monkey.PatchInstanceMethod(reflect.TypeOf(service), "JudgeBranchIsExists", func(service *Service, ctx context.Context, repoPath, branch string) (has bool, err error) {
		return true, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(service), "IdempotentCreateBranch", func(service *Service, ctx context.Context, repoPath, sourceBranch, newBranch string) error {
		return nil
	})

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetMergeBase", func(bdl *bundle.Bundle, userID string, req apistructs.GittarMergeBaseRequest) (*apistructs.Commit, error) {
		return &apistructs.Commit{
			ID: "1",
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListGittarCommit", func(bdl *bundle.Bundle, repo, ref, userID string, orgID string) (*apistructs.Commit, error) {
		return &apistructs.Commit{
			ID: "2",
		}, nil
	})

	type field struct {
		p *provider
	}
	type args struct {
		ctx                     context.Context
		devFlows                []db.DevFlow
		ruleNameBranchPolicyMap map[string]branchPolicy
		appMap                  map[uint64]apistructs.ApplicationDTO
		appInIssuePermissionMap map[uint64]bool
	}
	tests := []struct {
		name    string
		fields  field
		args    args
		want    map[string][]*pb.ChangeBranch
		want1   map[string]*apistructs.Commit
		wantErr bool
	}{
		{
			name: "test with error",
			fields: field{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: service,
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				devFlows: []db.DevFlow{
					{
						Model: db.Model{
							ID: fields.UUID{
								String: "157c8320-3755-402b-81ab-01d8bdd99512",
								Valid:  false,
							},
						},
						Scope:                db.Scope{},
						Operator:             db.Operator{},
						Branch:               "",
						IssueID:              0,
						FlowRuleName:         "",
						JoinTempBranchStatus: "",
						IsJoinTempBranch:     false,
					},
				},
				ruleNameBranchPolicyMap: nil,
				appMap:                  nil,
				appInIssuePermissionMap: nil,
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: field{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: service,
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				devFlows: []db.DevFlow{
					{
						Model: db.Model{
							ID: fields.UUID{
								String: "157c8320-3755-402b-81ab-01d8bdd99512",
								Valid:  false,
							},
						},
						Scope: db.Scope{
							AppID:   1,
							AppName: "erda",
						},
						Branch:               "feature/dop",
						IssueID:              1,
						FlowRuleName:         "DEV",
						JoinTempBranchStatus: "",
						IsJoinTempBranch:     true,
					},
					{
						Model: db.Model{
							ID: fields.UUID{
								String: "157c8320-3755-402b-81ab-01d8bdd99513",
								Valid:  false,
							},
						},
						Scope: db.Scope{
							AppID:   1,
							AppName: "erda",
						},
						Branch:               "feature/dop-test",
						IssueID:              1,
						FlowRuleName:         "TEST",
						JoinTempBranchStatus: "",
						IsJoinTempBranch:     true,
					},
				},
				ruleNameBranchPolicyMap: map[string]branchPolicy{
					"DEV": {
						tempBranch:   "next/dev",
						targetBranch: "master",
						sourceBranch: "master",
					},
				},
				appMap: map[uint64]apistructs.ApplicationDTO{
					1: {
						ID:          0,
						Name:        "erda",
						ProjectName: "erda",
					},
				},
				appInIssuePermissionMap: map[uint64]bool{
					1: true, 2: false,
				},
			},
			want: map[string][]*pb.ChangeBranch{
				"1next/dev": {
					{
						Commit: &pb.Commit{
							ID: "1",
						},
						BranchName: "feature/dop",
					},
				},
			},
			want1: map[string]*apistructs.Commit{
				"1next/dev": {ID: "2"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, got1, err := s.getAppTempBranchCommitAndChangeBranchListMap(tt.args.ctx, tt.args.devFlows, tt.args.ruleNameBranchPolicyMap, tt.args.appMap, tt.args.appInIssuePermissionMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAppTempBranchCommitAndChangeBranchListMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAppTempBranchCommitAndChangeBranchListMap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getAppTempBranchCommitAndChangeBranchListMap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestService_OperationMerge_Request_Validate(t *testing.T) {
	type args struct {
		req *pb.OperationMergeRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test with invalid parameter1",

			args: args{
				req: &pb.OperationMergeRequest{
					DevFlowID: "24aed4b7-7c49-479f-af84-8e1c93b00f64",
					Enable:    nil,
				},
			},
			wantErr: true,
		},
		{
			name: "test with invalid parameter2",
			args: args{
				req: &pb.OperationMergeRequest{
					DevFlowID: "",
					Enable:    wrapperspb.Bool(true),
				},
			},
			wantErr: true,
		},
		{
			name: "test with valid parameter",
			args: args{
				req: &pb.OperationMergeRequest{
					DevFlowID: "24aed4b7-7c49-479f-af84-8e1c93b00f64",
					Enable:    wrapperspb.Bool(true),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OperationMerge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestService_RejoinTempBranch(t *testing.T) {
	var dbClient *db.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "ListDevFlowByFlowRuleNameAndAppIDs", func(dbClient *db.Client, flowRuleName string, appIDs ...uint64) (fs []db.DevFlow, err error) {
		return []db.DevFlow{
			{
				Model: db.Model{
					ID: fields.UUID{
						String: "1",
					},
				},
				IsJoinTempBranch: false,
			},
			{
				Model: db.Model{
					ID: fields.UUID{
						String: "2",
					},
				},
				Branch:           "feature/dop",
				IsJoinTempBranch: true,
			},
			{
				Model: db.Model{
					ID: fields.UUID{
						String: "3",
					},
				},
				Branch:           "feature/pr/1",
				IsJoinTempBranch: true,
			},
		}, nil
	})

	defer monkey.UnpatchAll()

	var svc *Service
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "IdempotentDeleteBranch", func(svc *Service, ctx context.Context, repoPath, branch string) error {
		return nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "IdempotentCreateBranch", func(svc *Service, ctx context.Context, repoPath, sourceBranch, newBranch string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "MergeToTempBranch", func(svc *Service, ctx context.Context, tempBranch string, appID uint64, devFlow *db.DevFlow) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "IsMROpenedOrNotCreated", func(svc *Service, ctx context.Context, currentBranch, targetBranch string, appID uint64) (bool, error) {
		if currentBranch == "feature/dop" {
			return false, nil
		}
		if currentBranch == "feature/pr/1" {
			return true, nil
		}
		return false, fmt.Errorf("fail")
	})

	type field struct {
		p *provider
	}
	type args struct {
		ctx          context.Context
		tempBranch   string
		sourceBranch string
		targetBranch string
		devFlow      *db.DevFlow
		app          *apistructs.ApplicationDTO
	}
	tests := []struct {
		name    string
		fields  field
		args    args
		wantErr bool
	}{
		{
			name: "test with rejoin",
			fields: field{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
				},
			},
			args: args{
				ctx:          context.Background(),
				tempBranch:   "next/dev",
				sourceBranch: "develop",
				targetBranch: "master",
				devFlow: &db.DevFlow{
					Model:                db.Model{},
					Scope:                db.Scope{},
					Operator:             db.Operator{},
					Branch:               "",
					IssueID:              0,
					FlowRuleName:         "",
					JoinTempBranchStatus: "",
					IsJoinTempBranch:     false,
				},
				app: &apistructs.ApplicationDTO{
					ID:   1,
					Name: "erda",
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
			if err := s.RejoinTempBranch(tt.args.ctx, tt.args.tempBranch, tt.args.sourceBranch, tt.args.targetBranch, tt.args.devFlow, tt.args.app); (err != nil) != tt.wantErr {
				t.Errorf("RejoinTempBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteFlowNode(t *testing.T) {
	var dbClient *db.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetDevFlow", func(dbClient *db.Client, devFlowID string) (f *db.DevFlow, err error) {
		if devFlowID == "" {
			return nil, fmt.Errorf("empty devFlowID")
		}
		if devFlowID == "1" {
			return &db.DevFlow{
				IsJoinTempBranch: true,
				FlowRuleName:     "DEV",
			}, nil
		}
		if devFlowID == "2" {
			return &db.DevFlow{
				IsJoinTempBranch: false,
				FlowRuleName:     "DEV",
			}, nil
		}
		return nil, fmt.Errorf("record not found")
	})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "DeleteDevFlow", func(dbClient *db.Client, devFlowID string) error {
		return nil
	})

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp", func(d *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			Name:        "erda",
			ProjectName: "erda-project",
			ProjectID:   1,
		}, nil
	})

	var svc *Service
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "RejoinTempBranch", func(svc *Service, ctx context.Context, tempBranch, sourceBranch, targetBranch string, devFlow *db.DevFlow, app *apistructs.ApplicationDTO) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "IdempotentDeleteBranch", func(svc *Service, ctx context.Context, repoPath, branch string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "UpdateDevFlowAndDoRejoin", func(svc *Service, ctx context.Context, devFlow *db.DevFlow, app *apistructs.ApplicationDTO) error {
		return nil
	})

	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteFlowNodeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteFlowNodeResponse
		wantErr bool
	}{
		{
			name: "test with error1",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "",
					DeleteBranch: false,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with error2",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "3",
					DeleteBranch: false,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with isJoinTempBranch is false",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "2",
					DeleteBranch: false,
				},
			},
			want:    &pb.DeleteFlowNodeResponse{},
			wantErr: false,
		},
		{
			name: "test with isJoinTempBranch is true",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "1",
					DeleteBranch: false,
				},
			},
			want:    &pb.DeleteFlowNodeResponse{},
			wantErr: false,
		},
		{
			name: "test with delete branch1",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "1",
					DeleteBranch: true,
				},
			},
			want:    &pb.DeleteFlowNodeResponse{},
			wantErr: false,
		},
		{
			name: "test with delete branch2",
			fields: fields{
				p: &provider{
					dbClient:       dbClient,
					devFlowService: svc,
					DevFlowRule:    devFlowRuleForGetMock{},
					bdl:            bdl,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteFlowNodeRequest{
					DevFlowID:    "2",
					DeleteBranch: true,
				},
			},
			want:    &pb.DeleteFlowNodeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				p: tt.fields.p,
			}
			got, err := s.DeleteFlowNode(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFlowNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteFlowNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
