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

package executor

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/erda-project/erda-proto-go/core/rule/pb"
	"github.com/erda-project/erda/internal/core/rule/dao"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestExprExecutor_Exec(t *testing.T) {
	type args struct {
		r   *RuleConfig
		env map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			args: args{
				r: &RuleConfig{
					Code: "len(issue.content) > 2",
				},
				env: map[string]interface{}{
					"issue": map[string]interface{}{
						"content": "123",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExprExecutor{}
			got, err := e.Exec(tt.args.r, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExprExecutor.Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExprExecutor.Exec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExprExecutor_BuildRuleEnv(t *testing.T) {
	type args struct {
		req *pb.FireRequest
	}
	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "ListRuleSets",
		func(d *dao.DBClient, req *pb.ListRuleSetsRequest) ([]dao.RuleSet, int64, error) {
			return []dao.RuleSet{
				{
					ID:   "1",
					Code: "len(issue.content) > 2",
				},
			}, 1, nil
		},
	)

	defer p1.Unpatch()

	tests := []struct {
		name    string
		args    args
		want    *RuleEnv
		wantErr bool
	}{
		{
			args: args{
				req: &pb.FireRequest{
					EventType: "issue",
					Env: map[string]*structpb.Value{
						"content": structpb.NewStringValue("123"),
					},
					Scope:   "project",
					ScopeID: "1",
				},
			},
			want: &RuleEnv{
				Configs: []*RuleConfig{
					{
						RuleSetID: "1",
						Code:      "len(issue.content) > 2",
					},
				},
				Env: map[string]interface{}{
					"content": "123",
				},
			},
		},
		{
			args: args{
				req: &pb.FireRequest{
					EventType: "issue",
					Env: map[string]*structpb.Value{
						"content": structpb.NewStringValue("123"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExprExecutor{
				DB: db,
			}
			got, err := e.BuildRuleEnv(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExprExecutor.BuildRuleEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExprExecutor.BuildRuleEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
