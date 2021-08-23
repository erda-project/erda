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

package exception

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
)

func Test_exceptionService_GetExceptions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExceptionsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExceptionsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.exception.ExceptionService",
		//			`
		//erda.msp.apm.exception:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExceptionsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExceptionsResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
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
			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
			got, err := srv.GetExceptions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("exceptionService.GetExceptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("exceptionService.GetExceptions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_exceptionService_GetExceptionEventIds(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExceptionEventIdsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExceptionEventIdsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.exception.ExceptionService",
		//			`
		//erda.msp.apm.exception:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExceptionEventIdsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExceptionEventIdsResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
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
			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
			got, err := srv.GetExceptionEventIds(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("exceptionService.GetExceptionEventIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("exceptionService.GetExceptionEventIds() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_exceptionService_GetExceptionEvent(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExceptionEventRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExceptionEventResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.exception.ExceptionService",
		//			`
		//erda.msp.apm.exception:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExceptionEventRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExceptionEventResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
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
			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
			got, err := srv.GetExceptionEvent(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("exceptionService.GetExceptionEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("exceptionService.GetExceptionEvent() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
