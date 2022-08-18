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

	"bou.ke/monkey"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/jsonnet"
)

type mockidentity struct {
}

func (i *mockidentity) FindUsers(context.Context, *userpb.FindUsersRequest) (*userpb.FindUsersResponse, error) {
	return &userpb.FindUsersResponse{
		Data: []*userpb.User{
			{
				ID:    "1",
				Phone: "123",
			},
			{
				ID: "2",
			},
		},
	}, nil
}

func (i *mockidentity) GetUser(context.Context, *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	return nil, nil
}

func (i *mockidentity) FindUsersByKey(context.Context, *userpb.FindUsersByKeyRequest) (*userpb.FindUsersByKeyResponse, error) {
	return nil, nil
}

func Test_provider_getDingTalkConfig(t *testing.T) {
	var engine *jsonnet.Engine
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(engine), "EvaluateBySnippet",
		func(d *jsonnet.Engine, snippet string, configs []jsonnet.TLACodeConfig) (string, error) {
			return "{}", nil
		},
	)
	defer p1.Unpatch()

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
			args: args{
				param: &JsonnetParam{},
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
