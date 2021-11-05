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

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/log"
	"github.com/erda-project/erda/modules/orchestrator/services/resource"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/http/httpclient"
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

func TestAddon_prepareAttachment(t *testing.T) {
	type fields struct {
		db       *dbclient.DBClient
		bdl      *bundle.Bundle
		hc       *httpclient.HTTPClient
		encrypt  *encryption.EnvEncrypt
		resource *resource.Resource
		Logger   *log.DeployLogHelper
	}
	type args struct {
		addonInsRouting *dbclient.AddonInstanceRouting
		addonAttach     *dbclient.AddonAttachment
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "t1",
			fields: fields{},
			args:   args{
				addonInsRouting: &dbclient.AddonInstanceRouting{
					AddonName: "not mysql",
				},
				addonAttach:     nil,
			},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Addon{
				db:       tt.fields.db,
				bdl:      tt.fields.bdl,
				hc:       tt.fields.hc,
				encrypt:  tt.fields.encrypt,
				resource: tt.fields.resource,
				Logger:   tt.fields.Logger,
			}
			if got := a.prepareAttachment(tt.args.addonInsRouting, tt.args.addonAttach); got != tt.want {
				t.Errorf("prepareAttachment() = %v, want %v", got, tt.want)
			}
		})
	}
}