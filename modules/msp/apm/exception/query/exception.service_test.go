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
	eventpb "github.com/erda-project/erda-proto-go/core/monitor/event/pb"
	commonPb "github.com/erda-project/erda-proto-go/oap/common/pb"
	entitypb "github.com/erda-project/erda-proto-go/oap/entity/pb"
	oapPb "github.com/erda-project/erda-proto-go/oap/event/pb"
	"github.com/golang/mock/gomock"
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

//go:generate mockgen -destination=./mock_event_query_grpc.go -package query -source=../../../../../api/proto-go/core/monitor/event/pb/event_query_grpc.pb.go EventQueryServiceServer
//go:generate mockgen -destination=./mock_entity_query_grpc.go -package query -source=../../../../../api/proto-go/oap/entity/pb/entity_grpc.pb.go EntityServiceServer
func TestExceptionService_fetchErdaErrorFromES(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	entityGrpcServer := NewMockEntityServiceServer(ctrl)
	exceptionEntity := entitypb.Entity{
		Id:                 "error_exception/0f82da3be2e1c7070c269471fa7aa4a5",
		Type:               "error_exception",
		Key:                "0f82da3be2e1c7070c269471fa7aa4a5",
		Values:             nil,
		Labels:             nil,
		CreateTimeUnixNano: 1635845334935110000,
		UpdateTimeUnixNano: 1635851825720883700,
	}
	entityList := entitypb.EntityList{
		List:  []*entitypb.Entity{&exceptionEntity},
		Total: 1,
	}
	listEntitiesResponse := entitypb.ListEntitiesResponse{
		Data: &entityList,
	}
	entityGrpcServer.EXPECT().ListEntities(gomock.Any(), gomock.Any()).Return(&listEntitiesResponse, nil)

	eventGrpcServer := NewMockEventQueryServiceServer(ctrl)
	att := make(map[string]string)
	att["terminusKey"] = "fc1f8c074e46a9df505a15c1a94d62cc"
	spanEvent := oapPb.Event{
		EventID:      "335415fe-0c9f-4905-ab7f-434032a5c3ab",
		Severity:     "",
		Name:         "exception",
		Kind:         0,
		TimeUnixNano: 1635845334935610000,
		Relations:    &commonPb.Relation{
			TraceID:      "",
			ResID:        "0f82da3be2e1c7070c269471fa7aa4a5",
			ResType:      "exception",
			ResourceKeys: nil,
		},
		Attributes:   att,
		Message:      "",
	}
	eventsResult := eventpb.GetEventsResult{
		Items: []*oapPb.Event{&spanEvent},
	}
	eventsResponse := eventpb.GetEventsResponse{
		Data: &eventsResult,
	}
	eventGrpcServer.EXPECT().GetEvents(gomock.Any(), gomock.Any()).Return(&eventsResponse, nil)

	conditions := map[string]string{
		"terminusKey": "fc1f8c074e46a9df505a15c1a94d62cc",
	}
	entityReq := &entitypb.ListEntitiesRequest{
		Type:   "error_exception",
		Labels: conditions,
		Limit:  int64(1000),
	}

	items, err := fetchErdaErrorFromES(context.Background(), eventGrpcServer, entityGrpcServer, entityReq, 1635845334935, 1635851825720)

	if err != nil {
		t.Errorf("should not throw error")
	}
	if items == nil || len(items) != 1 {
		t.Errorf("assert result failed")
	}
		
}
