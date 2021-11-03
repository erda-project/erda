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
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	error_storage "github.com/erda-project/erda/modules/msp/apm/exception/erda-error/storage"
	event_storage "github.com/erda-project/erda/modules/msp/apm/exception/erda-event/storage"
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

func Test_exceptionService_fetchErdaEventFromES(t *testing.T) {
	//pbevent := &pb.ExceptionEvent{
	//	Id:             "Id",
	//	ExceptionID:    "ExceptionID",
	//	Metadata:       nil,
	//	RequestContext: nil,
	//	RequestHeaders: nil,
	//	RequestID:      "RequestID",
	//	Stacks:         nil,
	//	Tags:           nil,
	//	Timestamp:      0,
	//	RequestSampled: false,
	//}
	//
	//pberror := &pb.Exception{
	//	Id:               "Id",
	//	ClassName:        "ClassName",
	//	Method:           "Method",
	//	Type:             "Type",
	//	EventCount:       0,
	//	ExceptionMessage: "ExceptionMessage",
	//	File:             "File",
	//	ApplicationID:    "ApplicationID",
	//	RuntimeID:        "RuntimeID",
	//	ServiceName:      "ServiceName",
	//	ScopeID:          "ScopeID",
	//	CreateTime:       "CreateTime",
	//	UpdateTime:       "UpdateTime",
	//}

	erdaEvent := &exception.Erda_event{
		EventId:        "Id",
		Timestamp:      0,
		RequestId:      "RequestID",
		ErrorId:        "ExceptionID",
		Stacks:         nil,
		Tags:           nil,
		MetaData:       nil,
		RequestContext: nil,
		RequestHeaders: nil,
	}

	//erdaError := &exception.Erda_error{
	//	TerminusKey:   "ScopeID",
	//	ApplicationId: "ApplicationID",
	//	ServiceName:   "ServiceName",
	//	ErrorId:       "Id",
	//	Timestamp:     0,
	//	Tags:          nil,
	//}

	e1 := &errorEventListStorage{
		exceptionEvent: erdaEvent,
	}
	//e2 := &errorListStorage{
	//	exception: pberror,
	//}

	tests := []struct {
		name string
		ctx  context.Context
		//errorStorage error_storage.Storage
		eventStorage event_storage.Storage
		//errorSel     error_storage.Selector
		eventSel event_storage.Selector
		forward  bool
		limit    int
		want     []*exception.Erda_event
	}{{
		"case 1",
		context.TODO(),
		//e2,
		e1,
		//error_storage.Selector{},
		event_storage.Selector{},
		true,
		1,
		[]*exception.Erda_event{erdaEvent},
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got, err := fetchErdaEventFromES(tt.ctx, e1, &tt.eventSel, tt.forward, tt.limit); !reflect.DeepEqual(got, tt.want) || err != nil {
				t.Errorf("fetchErdaEventFromES() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exceptionService_fetchErdaErrorFromES(t *testing.T) {
	req := &pb.GetExceptionsRequest{
		StartTime: 0,
		EndTime:   0,
		ScopeID:   "",
	}

	resp := &pb.Exception{
		Id:               "Id",
		ClassName:        "",
		Method:           "",
		Type:             "",
		EventCount:       1,
		ExceptionMessage: "",
		File:             "",
		ApplicationID:    "",
		RuntimeID:        "",
		ServiceName:      "",
		ScopeID:          "ScopeID",
		CreateTime:       "1970-01-01 08:00:00",
		UpdateTime:       "1970-01-01 08:00:00",
	}

	erdaEvent := &exception.Erda_event{
		EventId:        "Id",
		Timestamp:      0,
		RequestId:      "RequestID",
		ErrorId:        "ExceptionID",
		Stacks:         nil,
		Tags:           nil,
		MetaData:       nil,
		RequestContext: nil,
		RequestHeaders: nil,
	}

	erdaError := &exception.Erda_error{
		TerminusKey:   "ScopeID",
		ApplicationId: "ApplicationID",
		ServiceName:   "ServiceName",
		ErrorId:       "Id",
		Timestamp:     0,
		Tags:          nil,
	}

	errorEventListStorage := &errorEventListStorage{
		exceptionEvent: erdaEvent,
	}
	errorListStorage := &errorListStorage{
		exception: erdaError,
	}

	tests := []struct {
		name         string
		ctx          context.Context
		errorStorage error_storage.Storage
		eventStorage event_storage.Storage
		errorSel     error_storage.Selector
		eventSel     event_storage.Selector
		forward      bool
		limit        int
		want         []*pb.Exception
	}{{
		"case 1",
		context.TODO(),
		//e2,
		errorListStorage,
		errorEventListStorage,
		error_storage.Selector{},
		event_storage.Selector{},
		true,
		1,
		[]*pb.Exception{resp},
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got, err := fetchErdaErrorFromES(tt.ctx, errorListStorage, errorEventListStorage, req, tt.forward, tt.limit); !reflect.DeepEqual(got, tt.want) || err != nil {
				t.Errorf("fetchErdaErrorFromES() = %v, want %v", got, tt.want)
			}
		})
	}
}
