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

package accesskey

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
)

func Test_accessKeyService_QueryAccessKeys(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		req *pb.QueryAccessKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.QueryAccessKeysResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.QueryAccessKeysRequest{},
			},
			want: &pb.QueryAccessKeysResponse{
				Data: []*pb.AccessKeysItem{
					{
						Id:          "aaa",
						AccessKey:   "xxx",
						SecretKey:   "yyy",
						Status:      pb.StatusEnum_ACTIVATE,
						SubjectType: pb.SubjectTypeEnum_MICRO_SERVICE,
						Subject:     "1",
						Description: "xxx",
						CreatedAt:   timestamppb.New(_mockTime),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.QueryAccessKeysRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &accessKeyService{
				p: tt.fields.p,
			}
			got, err := s.QueryAccessKeys(context.TODO(), tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryAccessKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QueryAccessKeys() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accessKeyService_GetAccessKey(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetAccessKeyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetAccessKeyResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.GetAccessKeyRequest{},
			},
			want: &pb.GetAccessKeyResponse{
				Data: &pb.AccessKeysItem{
					Id:          "aaa",
					AccessKey:   "xxx",
					SecretKey:   "yyy",
					Status:      pb.StatusEnum_ACTIVATE,
					SubjectType: pb.SubjectTypeEnum_MICRO_SERVICE,
					Subject:     "1",
					Description: "xxx",
					CreatedAt:   timestamppb.New(_mockTime),
				},
			},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.GetAccessKeyRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &accessKeyService{
				p: tt.fields.p,
			}
			got, err := s.GetAccessKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccessKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccessKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accessKeyService_CreateAccessKeys(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.CreateAccessKeyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CreateAccessKeyResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.CreateAccessKeyRequest{},
			},
			want: &pb.CreateAccessKeyResponse{
				Data: &pb.AccessKeysItem{
					Id:          "aaa",
					AccessKey:   "xxx",
					SecretKey:   "yyy",
					Status:      pb.StatusEnum_ACTIVATE,
					SubjectType: pb.SubjectTypeEnum_MICRO_SERVICE,
					Subject:     "1",
					Description: "xxx",
					CreatedAt:   timestamppb.New(_mockTime),
				},
			},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.CreateAccessKeyRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &accessKeyService{
				p: tt.fields.p,
			}
			got, err := s.CreateAccessKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccessKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAccessKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accessKeyService_UpdateAccessKeys(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.UpdateAccessKeyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UpdateAccessKeyResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.UpdateAccessKeyRequest{},
			},
			want:    &pb.UpdateAccessKeyResponse{},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.UpdateAccessKeyRequest{},
			},
			want:    &pb.UpdateAccessKeyResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &accessKeyService{
				p: tt.fields.p,
			}
			got, err := s.UpdateAccessKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAccessKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateAccessKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accessKeyService_DeleteAccessKeys(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteAccessKeyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteAccessKeyResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.DeleteAccessKeyRequest{},
			},
			want:    &pb.DeleteAccessKeyResponse{},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.DeleteAccessKeyRequest{},
			},
			want:    &pb.DeleteAccessKeyResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &accessKeyService{
				p: tt.fields.p,
			}
			got, err := s.DeleteAccessKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccessKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteAccessKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
