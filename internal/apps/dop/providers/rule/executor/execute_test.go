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
	"testing"

	"github.com/erda-project/erda-proto-go/dop/rule/pb"

	"github.com/erda-project/erda/internal/apps/dop/providers/rule/actions/api"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/db"
)

type MockAPI struct {
}

func (a *MockAPI) Send(api *api.API) (string, error) {
	return "ok", nil
}

func TestExecutor_Do(t *testing.T) {
	type args struct {
		content map[string]interface{}
		config  *RuleConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{
				content: map[string]interface{}{
					"k": "v",
				},
				config: &RuleConfig{},
			},
			want: "no valid action nodes",
		},
		{
			args: args{
				content: map[string]interface{}{
					"k": "v",
				},
				config: &RuleConfig{
					Action: db.ActionParams{
						Nodes: []*pb.ActionNode{{
							Snippet: "123",
							Type:    "api",
						}},
					},
				},
			},
			want: "ok",
		},
	}
	e := &Executor{
		API: &MockAPI{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Do(tt.args.content, tt.args.config)
			if got != tt.want {
				t.Errorf("Executor.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}
