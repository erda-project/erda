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

package dataview

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/dataview/pb"
)

func Test_dataViewService_ListSystemViews(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListSystemViewsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListSystemViewsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.ListSystemViewsRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.ListSystemViewsResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.ListSystemViews(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.ListSystemViews() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.ListSystemViews() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_GetSystemView(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetSystemViewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetSystemViewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetSystemViewRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetSystemViewResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.GetSystemView(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.GetSystemView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.GetSystemView() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_CreateCustomView(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateCustomViewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateCustomViewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.CreateCustomViewRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.CreateCustomViewResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.CreateCustomView(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.CreateCustomView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.CreateCustomView() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_ListCustomViews(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListCustomViewsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListCustomViewsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.ListCustomViewsRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.ListCustomViewsResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.ListCustomViews(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.ListCustomViews() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.ListCustomViews() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_GetCustomView(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetCustomViewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetCustomViewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetCustomViewRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetCustomViewResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.GetCustomView(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.GetCustomView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.GetCustomView() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_UpdateCustomView(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.UpdateCustomViewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.UpdateCustomViewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.UpdateCustomViewRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.UpdateCustomViewResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.UpdateCustomView(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.UpdateCustomView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.UpdateCustomView() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_dataViewService_DeleteCustomView(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.DeleteCustomViewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.DeleteCustomViewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.dataview.DataViewService",
		// 			`
		// erda.core.monitor.dataview:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.DeleteCustomViewRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.DeleteCustomViewResponse{
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
			srv := hub.Service(tt.service).(pb.DataViewServiceServer)
			got, err := srv.DeleteCustomView(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataViewService.DeleteCustomView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("dataViewService.DeleteCustomView() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
