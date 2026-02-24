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

package util

import (
	"reflect"
	"testing"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/apistructs"
)

func TestDensensitize(t *testing.T) {
	type args struct {
		IDs             []string
		b               []*commonpb.UserInfo
		needDesensitize bool
	}
	tests := []struct {
		name string
		args args
		want map[string]apistructs.UserInfo
	}{
		{
			args: args{
				IDs: []string{"1"},
				b: []*commonpb.UserInfo{
					{
						Id:    "1",
						Email: "test@test.com",
					},
				},
				needDesensitize: true,
			},
			want: map[string]apistructs.UserInfo{
				"1": {
					Email: "te*t@test.com",
				},
			},
		},
		{
			args: args{
				IDs: []string{"1", "2"},
				b: []*commonpb.UserInfo{
					{
						Id:    "1",
						Email: "test@test.com",
					},
				},
			},
			want: map[string]apistructs.UserInfo{
				"1": {
					ID:    "1",
					Email: "te*t@test.com",
				},
				"2": {
					ID:   "2",
					Name: "用户已注销",
					Nick: "用户已注销",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Densensitize(tt.args.IDs, tt.args.b, tt.args.needDesensitize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Densensitize() = %v, want %v", got, tt.want)
			}
		})
	}
}
