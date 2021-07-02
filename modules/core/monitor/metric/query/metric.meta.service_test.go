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

func Test_metricMetaService_ListMetricNames(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListMetricNamesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListMetricNamesResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricMetaService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.ListMetricNamesRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.ListMetricNamesResponse{
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
			srv := hub.Service(tt.service).(pb.MetricMetaServiceServer)
			got, err := srv.ListMetricNames(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricMetaService.ListMetricNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricMetaService.ListMetricNames() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricMetaService_ListMetricMeta(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListMetricMetaRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListMetricMetaResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricMetaService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.ListMetricMetaRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.ListMetricMetaResponse{
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
			srv := hub.Service(tt.service).(pb.MetricMetaServiceServer)
			got, err := srv.ListMetricMeta(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricMetaService.ListMetricMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricMetaService.ListMetricMeta() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricMetaService_ListMetricGroups(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListMetricGroupsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListMetricGroupsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricMetaService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.ListMetricGroupsRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.ListMetricGroupsResponse{
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
			srv := hub.Service(tt.service).(pb.MetricMetaServiceServer)
			got, err := srv.ListMetricGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricMetaService.ListMetricGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricMetaService.ListMetricGroups() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_metricMetaService_GetMetricGroup(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetMetricGroupRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetMetricGroupResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.metric.MetricMetaService",
		// 			`
		// erda.core.monitor.metric:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetMetricGroupRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetMetricGroupResponse{
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
			srv := hub.Service(tt.service).(pb.MetricMetaServiceServer)
			got, err := srv.GetMetricGroup(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricMetaService.GetMetricGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("metricMetaService.GetMetricGroup() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
