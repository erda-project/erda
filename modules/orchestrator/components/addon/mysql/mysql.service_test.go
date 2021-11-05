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

package mysql

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

type MockKMS struct {
}

func (k *MockKMS) CreateKey() (*kmstypes.CreateKeyResponse, error) {
	return &kmstypes.CreateKeyResponse{KeyMetadata: kmstypes.KeyMetadata{
		KeyID: "123",
	}}, nil
}

func (k *MockKMS) Encrypt(plaintext, keyID string) (*kmstypes.EncryptResponse, error) {
	return &kmstypes.EncryptResponse{
		KeyID:            "123",
		CiphertextBase64: "***",
	}, nil
}

func (k *MockKMS) Decrypt(ciphertext, keyID string) (*kmstypes.DecryptResponse, error) {
	return &kmstypes.DecryptResponse{
		PlaintextBase64: "MjIy",
	}, nil
}

func Test_mysqlService_ToDTO(t *testing.T) {
	type fields struct {
		logger logs.Logger
		kms    KMSWrapper
		db     *dbclient.DBClient
	}
	type args struct {
		acc     *dbclient.MySQLAccount
		decrypt bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *pb.MySQLAccount
	}{
		{
			name: "t1",
			fields: fields{
				kms: &MockKMS{},
			},
			args: args{
				acc: &dbclient.MySQLAccount{
					ID:                "1",
					CreatedAt:         time.Time{},
					UpdatedAt:         time.Time{},
					Username:          "mysql-123",
					Password:          "MjIy",
					KMSKey:            "123",
					InstanceID:        "inst1",
					RoutingInstanceID: "r_inst2",
					Creator:           "321",
					IsDeleted:         false,
				},
				decrypt: true,
			},
			want: &pb.MySQLAccount{
				Id:         "1",
				InstanceId: "r_inst2",
				Creator:    "321",
				CreateAt:   timestamppb.New(time.Time{}),
				Username:   "mysql-123",
				Password:   "222",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mysqlService{
				logger: tt.fields.logger,
				kms:    tt.fields.kms,
				db:     tt.fields.db,
			}
			if got := s.ToDTO(tt.args.acc, tt.args.decrypt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDTO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mysqlService_getConnConfig(t *testing.T) {
	type fields struct {
		logger logs.Logger
		kms    KMSWrapper
		db     *dbclient.DBClient
	}
	type args struct {
		ins *dbclient.AddonInstance
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *connConfig
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				kms: &MockKMS{},
			},
			args:    args{
				ins: &dbclient.AddonInstance{
					Config: `{"MYSQL_HOST":"hhh","MYSQL_PORT":"ppp","MYSQL_USERNAME":"uuu","MYSQL_PASSWORD":"MjIy"}`,
				},
			},
			want:    &connConfig{
				Host: "hhh",
				Port: "ppp",
				User: "uuu",
				Pass: "222",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mysqlService{
				logger: tt.fields.logger,
				kms:    tt.fields.kms,
				db:     tt.fields.db,
			}
			got, err := s.getConnConfig(tt.args.ins)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConnConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConnConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
