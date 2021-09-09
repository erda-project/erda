// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
					Content:    "goodbye world",
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
					Content:    "goodbye world",
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
					Content:    "goodbye world",
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
