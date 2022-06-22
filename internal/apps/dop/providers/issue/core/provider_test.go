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

package core

import (
	"context"
	"testing"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

func TestScopeID(t *testing.T) {
	type args struct {
		ctx context.Context
		req interface{}
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{ctx, &pb.GetIssueStatesRequest{ProjectID: 1}},
			want: "1",
		},
		{
			args: args{ctx, &pb.DeleteIssueStateRequest{ProjectID: 1}},
			want: "1",
		},
		{
			args:    args{ctx, &pb.DeleteIssueRequest{}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScopeID(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScopeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ScopeID() = %v, want %v", got, tt.want)
			}
		})
	}
}
