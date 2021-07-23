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
	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
)

func Test_logQueryService_GetLog(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetLogRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetLogResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.monitor.log.query.LogQueryService",
			`
erda.core.monitor.log.query:
`,
			args{
				context.TODO(),
				&pb.GetLogRequest{
					// TODO: setup fields
				},
			},
			&pb.GetLogResponse{
				// TODO: setup fields.
			},
			false,
		},
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
			srv := hub.Service(tt.service).(pb.LogQueryServiceServer)
			got, err := srv.GetLog(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logQueryService.GetLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logQueryService.GetLog() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logQueryService_GetLogByRuntime(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetLogByRuntimeRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetLogByRuntimeResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.monitor.log.query.LogQueryService",
			`
erda.core.monitor.log.query:
`,
			args{
				context.TODO(),
				&pb.GetLogByRuntimeRequest{
					// TODO: setup fields
				},
			},
			&pb.GetLogByRuntimeResponse{
				// TODO: setup fields.
			},
			false,
		},
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
			srv := hub.Service(tt.service).(pb.LogQueryServiceServer)
			got, err := srv.GetLogByRuntime(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logQueryService.GetLogByRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logQueryService.GetLogByRuntime() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logQueryService_GetLogByOrganization(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetLogByOrganizationRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetLogByOrganizationResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.monitor.log.query.LogQueryService",
			`
erda.core.monitor.log.query:
`,
			args{
				context.TODO(),
				&pb.GetLogByOrganizationRequest{
					// TODO: setup fields
				},
			},
			&pb.GetLogByOrganizationResponse{
				// TODO: setup fields.
			},
			false,
		},
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
			srv := hub.Service(tt.service).(pb.LogQueryServiceServer)
			got, err := srv.GetLogByOrganization(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logQueryService.GetLogByOrganization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logQueryService.GetLogByOrganization() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
