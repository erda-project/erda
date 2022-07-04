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

package devflowrule

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/internal/apps/dop/providers/devflowrule/db"
)

func Test_provider_GetFlowByRule(t *testing.T) {
	type fields struct {
		dbClient *db.Client
	}
	type args struct {
		ctx     context.Context
		request GetFlowByRuleRequest
	}
	dbClient := &db.Client{}
	monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetDevFlowRuleByProjectID", func(dbClient *db.Client, proID uint64) (fs *db.DevFlowRule, err error) {
		return &db.DevFlowRule{
			Flows: db.JSON(`[
    {
        "name": "PROD",
        "targetBranch": "master,support/*",
        "artifact": "stable",
        "environment": "PROD"
    },
    {
        "name": "STAGING",
        "targetBranch": "release/*,hotfix/*",
        "artifact": "rc",
        "environment": "STAGING"
    },
    {
        "name": "DEV",
        "targetBranch": "feature/*",
        "artifact": "alpha",
        "environment": "DEV"
    },
    {
        "name": "TEST",
        "targetBranch": "develop",
        "artifact": "alpha",
        "environment": "DEV"
    }
]`),
			BranchPolicies: db.JSON(
				`[
    {
        "branch": "feature/*",
        "branchType": "multi_branch",
        "policy": {
            "sourceBranch": "develop",
            "currentBranch": "feature/*",
            "tempBranch": "next/develop",
            "branchType": "",
            "targetBranch": {
                "mergeRequest": "develop",
                "cherryPick": ""
            }
        }
    },
    {
        "branch": "master,support/*",
        "branchType": "single_branch",
        "policy": null
    },
    {
        "branch": "develop",
        "branchType": "single_branch",
        "policy": null
    },
    {
        "branch": "release/*,hotfix/*",
        "branchType": "single_branch",
        "policy": null
    }
]`,
			),
		}, nil
	})
	defer monkey.UnpatchAll()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:   "test with GetFlowByRule",
			fields: fields{dbClient: dbClient},
			args: args{
				ctx: context.Background(),
				request: GetFlowByRuleRequest{
					ProjectID:     1,
					BranchType:    "multi_branch",
					CurrentBranch: "feature/123",
					SourceBranch:  "develop",
				},
			},
			want:    "next/develop",
			wantErr: false,
		},
		{
			name:   "test with nil1",
			fields: fields{dbClient: dbClient},
			args: args{
				ctx: context.Background(),
				request: GetFlowByRuleRequest{
					ProjectID:     1,
					BranchType:    "multi_branch",
					CurrentBranch: "feat/123",
					SourceBranch:  "master",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test with nil2",
			fields: fields{dbClient: dbClient},
			args: args{
				ctx: context.Background(),
				request: GetFlowByRuleRequest{
					ProjectID:     1,
					BranchType:    "single_branch",
					CurrentBranch: "feat/123",
					SourceBranch:  "master",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				dbClient: tt.fields.dbClient,
			}
			got, err := p.GetFlowByRule(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFlowByRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want != nil {
					t.Errorf("fail")
				}
			} else if got.BranchPolicy.Policy.TempBranch != tt.want.(string) {
				t.Errorf("GetFlowByRule() got = %v, want %v", got.BranchPolicy.Policy.TempBranch, tt.want)
			}
		})
	}
}
