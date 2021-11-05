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

package addon

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func Test_buildMySQLAccount(t *testing.T) {
	type args struct {
		addonIns        *dbclient.AddonInstance
		addonInsRouting *dbclient.AddonInstanceRouting
		extra           *dbclient.AddonInstanceExtra
		operator        string
	}
	tests := []struct {
		name string
		args args
		want *dbclient.MySQLAccount
	}{
		{
			name: "t1",
			args: args{
				addonIns: &dbclient.AddonInstance{
					ID:     "111",
					KmsKey: "123",
				},
				addonInsRouting: &dbclient.AddonInstanceRouting{
					ID: "222",
				},
				extra: &dbclient.AddonInstanceExtra{
					Value: "pass",
				},
				operator: "333",
			},
			want: &dbclient.MySQLAccount{
				KMSKey:            "123",
				Username:          "mysql",
				Password:          "pass",
				InstanceID:        "111",
				RoutingInstanceID: "222",
				Creator:           "333",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMySQLAccount(tt.args.addonIns, tt.args.addonInsRouting, tt.args.extra, tt.args.operator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildMySQLAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
