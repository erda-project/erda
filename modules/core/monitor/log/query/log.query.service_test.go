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

package query

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
)

func Test_logQueryService_GetLog(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetLogRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetLogResponse
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				p: mockProvider(),
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogRequest{
					Id:        "aaa",
					Source:    "container",
					Stream:    "stdout",
					RequestId: "",
					Start:     1604880001000000000,
					End:       1604880002000000000,
					Count:     -200,
				},
			},
			want: &pb.GetLogResponse{Lines: []*pb.LogItem{
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880001000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880002000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &logQueryService{
				p: tt.fields.p,
			}
			got, err := s.GetLog(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLog() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logQueryService_GetLogByRuntime(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetLogByRuntimeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetLogByRuntimeResponse
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				p: mockProvider(),
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByRuntimeRequest{
					Id:            "aaa",
					Source:        "container",
					Stream:        "stdout",
					RequestId:     "",
					Start:         1604880001000000000,
					End:           1604880002000000000,
					Count:         -200,
					ApplicationId: "app-1",
				},
			},
			want: &pb.GetLogByRuntimeResponse{Lines: []*pb.LogItem{
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880001000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880002000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
			}},
			wantErr: false,
		},
		{
			name: "normal but error",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						errorTrigger: true,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByRuntimeRequest{
					Id:            "aaa",
					Source:        "container",
					Stream:        "stdout",
					RequestId:     "",
					Start:         1604880001000000000,
					End:           1604880002000000000,
					Count:         -200,
					ApplicationId: "app-1",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal but empty",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						emptyResult: true,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByRuntimeRequest{
					Id:            "aaa",
					Source:        "container",
					Stream:        "stdout",
					RequestId:     "",
					Start:         1604880001000000000,
					End:           1604880002000000000,
					Count:         -200,
					ApplicationId: "app-1",
				},
			},
			want:    &pb.GetLogByRuntimeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &logQueryService{
				p: tt.fields.p,
			}
			got, err := s.GetLogByRuntime(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogByRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLogByRuntime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logQueryService_GetLogByOrganization(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetLogByOrganizationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetLogByOrganizationResponse
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				p: mockProvider(),
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByOrganizationRequest{
					Id:          "aaa",
					Source:      "container",
					Stream:      "stdout",
					RequestId:   "",
					Start:       1604880001000000000,
					End:         1604880002000000000,
					Count:       -200,
					ClusterName: "cluster-1",
				},
			},
			want: &pb.GetLogByOrganizationResponse{Lines: []*pb.LogItem{
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880001000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
				{
					Id:         "aaa",
					Source:     "container",
					Stream:     "stdout",
					TimeBucket: "1604880000000000000",
					Timestamp:  "1604880002000000000",
					Offset:     "11",
					Content:    "hello world",
					Level:      "INFO",
					RequestId:  "",
				},
			}},
			wantErr: false,
		},
		{
			name: "normal but error",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						errorTrigger: true,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByOrganizationRequest{
					Id:          "aaa",
					Source:      "container",
					Stream:      "stdout",
					RequestId:   "",
					Start:       1604880001000000000,
					End:         1604880002000000000,
					Count:       -200,
					ClusterName: "cluster-1",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal but empty",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						emptyResult: true,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: &pb.GetLogByOrganizationRequest{
					Id:          "aaa",
					Source:      "container",
					Stream:      "stdout",
					RequestId:   "",
					Start:       1604880001000000000,
					End:         1604880002000000000,
					Count:       -200,
					ClusterName: "cluster-1",
				},
			},
			want:    &pb.GetLogByOrganizationResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &logQueryService{
				p: tt.fields.p,
			}
			got, err := s.GetLogByOrganization(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogByOrganization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLogByOrganization() got = %v, want %v", got, tt.want)
			}
		})
	}
}
