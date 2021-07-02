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
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

func Test_metricService_QueryWithInfluxFormat(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryWithInfluxFormatRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryWithInfluxFormatResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.QueryWithInfluxFormatRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.QueryWithInfluxFormatResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.MetricServiceServer)
			got, err := srv.QueryWithInfluxFormat(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricService.QueryWithInfluxFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricService.QueryWithInfluxFormat() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricService_SearchWithInfluxFormat(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryWithInfluxFormatRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryWithInfluxFormatResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.QueryWithInfluxFormatRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.QueryWithInfluxFormatResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.MetricServiceServer)
			got, err := srv.SearchWithInfluxFormat(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricService.SearchWithInfluxFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricService.SearchWithInfluxFormat() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricService_QueryWithTableFormat(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryWithTableFormatRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryWithTableFormatResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.QueryWithTableFormatRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.QueryWithTableFormatResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.MetricServiceServer)
			got, err := srv.QueryWithTableFormat(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricService.QueryWithTableFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricService.QueryWithTableFormat() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricService_SearchWithTableFormat(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryWithTableFormatRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryWithTableFormatResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.QueryWithTableFormatRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.QueryWithTableFormatResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.MetricServiceServer)
			got, err := srv.SearchWithTableFormat(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricService.SearchWithTableFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricService.SearchWithTableFormat() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
