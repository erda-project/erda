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

package log_service

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/apm/log-service/pb"
)

func Test_logService_HistogramAggregation(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.HistogramAggregationRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.HistogramAggregationResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.apm.log_service.LogService",
			`
erda.msp.apm.log_service:
`,
			args{
				context.TODO(),
				&pb.HistogramAggregationRequest{
					// TODO: setup fields
				},
			},
			&pb.HistogramAggregationResponse{
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
			srv := hub.Service(tt.service).(pb.LogServiceServer)
			got, err := srv.HistogramAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logService.HistogramAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logService.HistogramAggregation() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logService_BucketAggregation(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.BucketAggregationRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.BucketAggregationResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.apm.log_service.LogService",
			`
erda.msp.apm.log_service:
`,
			args{
				context.TODO(),
				&pb.BucketAggregationRequest{
					// TODO: setup fields
				},
			},
			&pb.BucketAggregationResponse{
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
			srv := hub.Service(tt.service).(pb.LogServiceServer)
			got, err := srv.BucketAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logService.BucketAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logService.BucketAggregation() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logService_PagedSearch(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PagedSearchRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PagedSearchResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.apm.log_service.LogService",
			`
erda.msp.apm.log_service:
`,
			args{
				context.TODO(),
				&pb.PagedSearchRequest{
					// TODO: setup fields
				},
			},
			&pb.PagedSearchResponse{
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
			srv := hub.Service(tt.service).(pb.LogServiceServer)
			got, err := srv.PagedSearch(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logService.PagedSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logService.PagedSearch() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logService_SequentialSearch(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.SequentialSearchRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.SequentialSearchResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.apm.log_service.LogService",
			`
erda.msp.apm.log_service:
`,
			args{
				context.TODO(),
				&pb.SequentialSearchRequest{
					// TODO: setup fields
				},
			},
			&pb.SequentialSearchResponse{
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
			srv := hub.Service(tt.service).(pb.LogServiceServer)
			got, err := srv.SequentialSearch(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logService.SequentialSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logService.SequentialSearch() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_logService_GetFieldSettings(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetFieldSettingsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetFieldSettingsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.apm.log_service.LogService",
			`
erda.msp.apm.log_service:
`,
			args{
				context.TODO(),
				&pb.GetFieldSettingsRequest{
					// TODO: setup fields
				},
			},
			&pb.GetFieldSettingsResponse{
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
			srv := hub.Service(tt.service).(pb.LogServiceServer)
			got, err := srv.GetFieldSettings(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("logService.GetFieldSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("logService.GetFieldSettings() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
