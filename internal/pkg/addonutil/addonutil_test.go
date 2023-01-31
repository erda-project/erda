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

package addonutil

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestTransAddonName(t *testing.T) {
	type args struct {
		addonName string
	}

	cases := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				addonName: apistructs.AddonESAlias,
			},
			want: apistructs.AddonES,
		},
		{
			name: "case 2",
			args: args{
				addonName: apistructs.AddonZookeeperAlias,
			},
			want: apistructs.AddonZookeeper,
		},
		{
			name: "case 3",
			args: args{
				addonName: apistructs.AddonConfigCenterAlias,
			},
			want: apistructs.AddonConfigCenter,
		},
		{
			name: "case 4",
			args: args{
				addonName: apistructs.AddonRedis,
			},
			want: apistructs.AddonRedis,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := TransAddonName(tt.args.addonName); got != tt.want {
				t.Errorf("TransAddonName() = %v, want %v", got, tt.want)
			}
		})
	}
}
