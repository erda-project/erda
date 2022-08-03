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

package kratos

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
)

func Test_provider_GetUser(t *testing.T) {
	var p *provider
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertUserIDs", func(_ *provider, ids []string) ([]string, map[string]string, error) {
		return []string{"1"}, map[string]string{"1": "a"}, nil
	})
	monkey.Patch(getUserByID, func(kratosPrivateAddr string, userID string) (*common.User, error) {
		return &common.User{ID: "1"}, nil
	})
	defer monkey.UnpatchAll()
	type args struct {
		ctx context.Context
		req *pb.GetUserRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.GetUserResponse
		wantErr bool
	}{
		{
			args: args{
				ctx: context.Background(),
				req: &pb.GetUserRequest{
					UserID: "1",
				},
			},
			want: &pb.GetUserResponse{
				Data: &pb.User{
					ID: "a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			got, err := p.GetUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("provider.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
