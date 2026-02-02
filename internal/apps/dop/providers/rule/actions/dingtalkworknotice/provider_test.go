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

package dingtalkworknotice

import (
	"context"
	"reflect"
	"testing"

	gojsonnet "github.com/google/go-jsonnet"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/jsonnet"
)

type mockidentity struct {
	userpb.UnimplementedUserServiceServer
}

func (i *mockidentity) FindUsers(context.Context, *userpb.FindUsersRequest) (*userpb.FindUsersResponse, error) {
	return &userpb.FindUsersResponse{
		Data: []*commonpb.UserInfo{
			{
				Id:    "1",
				Phone: "123",
			},
			{
				Id: "2",
			},
		},
	}, nil
}

func Test_provider_getDingTalkConfig(t *testing.T) {
	engine := &jsonnet.Engine{
		JsonnetVM: gojsonnet.MakeVM(),
	}

	type args struct {
		param *JsonnetParam
	}
	tests := []struct {
		name    string
		args    args
		want    *DingTalkConfig
		wantErr bool
	}{
		{
			name: "snippet returns empty object, identity returns users with phone",
			args: args{
				param: &JsonnetParam{
					Snippet: "{}",
					TLARaw:  nil,
				},
			},
			want: &DingTalkConfig{
				Users: []string{"123"},
			},
		},
	}

	p := &provider{
		Identity:       &mockidentity{},
		TemplateParser: engine,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.getDingTalkConfig(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.getDingTalkConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("provider.getDingTalkConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
