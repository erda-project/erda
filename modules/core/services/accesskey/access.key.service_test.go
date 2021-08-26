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

	"github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
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
		req *pb.GetAccessKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetAccessKeysResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.GetAccessKeysRequest{},
			},
			want: &pb.GetAccessKeysResponse{
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
				req: &pb.GetAccessKeysRequest{},
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
		req *pb.CreateAccessKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CreateAccessKeysResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.CreateAccessKeysRequest{},
			},
			want: &pb.CreateAccessKeysResponse{
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
				req: &pb.CreateAccessKeysRequest{},
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
			got, err := s.CreateAccessKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccessKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAccessKeys() got = %v, want %v", got, tt.want)
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
		req *pb.UpdateAccessKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UpdateAccessKeysResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.UpdateAccessKeysRequest{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.UpdateAccessKeysRequest{},
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
			got, err := s.UpdateAccessKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAccessKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateAccessKeys() got = %v, want %v", got, tt.want)
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
		req *pb.DeleteAccessKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteAccessKeysResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				p: &provider{dao: &mockDao{}},
			},
			args: args{
				req: &pb.DeleteAccessKeysRequest{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				p: &provider{dao: &mockDao{errorTrigger: true}},
			},
			args: args{
				req: &pb.DeleteAccessKeysRequest{},
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
			got, err := s.DeleteAccessKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccessKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteAccessKeys() got = %v, want %v", got, tt.want)
			}
		})
	}
}
