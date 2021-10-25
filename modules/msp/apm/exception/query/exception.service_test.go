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

//import (
//	"bou.ke/monkey"
//	context "context"
//	reflect "reflect"
//	testing "testing"
//
//	servicehub "github.com/erda-project/erda-infra/base/servicehub"
//	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
//)
//
//func Test_exceptionService_GetExceptions(t *testing.T) {
//	type args struct {
//		ctx context.Context
//		req *pb.GetExceptionsRequest
//	}
//	tests := []struct {
//		name     string
//		service  string
//		config   string
//		args     args
//		wantResp *pb.GetExceptionsResponse
//		wantErr  bool
//	}{
//		// TODO: Add test cases.
//		{
//			"case 1",
//			"erda.msp.apm.exception.ExceptionService",
//			`erda.msp.apm.exception.query:`,
//			args{
//				context.TODO(),
//				&pb.GetExceptionsRequest{
//					// TODO: setup fields
//					StartTime: 1633790981924,
//					EndTime:   1635790981924,
//					ScopeID:   "fc1f8c074e46a9df505a15c1a94d62cc",
//				},
//			},
//			&pb.GetExceptionsResponse{
//				// TODO: setup fields.
//				Data: []*pb.Exception{
//					{
//						Id:               "cd41c7109d96edb62f3bb05380e78ab5",
//						ClassName:        "org.apache.dubbo.remoting.transport.netty4.NettyClient",
//						Method:           "doConnect",
//						Type:             "org.apache.dubbo.remoting.RemotingException",
//						EventCount:       17,
//						ExceptionMessage: "client(url: dubbo://30.43.56.70:20880/io.terminus.demo.rpc.DubboService?anyhost=true&application=apm-demo-api&category=providers&check=false&codec=dubbo&deprecated=false&dubbo=2.0.2&dynamic=true&generic=false&heartbeat=60000&init=false&interface=io.terminus.demo.rpc.DubboService&metadata-type=remote&methods=mysqlUsers,redisGet,httpRequest,hello,error&path=io.terminus.demo.rpc.DubboService&pid=1&protocol=dubbo&qos.enable=false&register.ip=10.112.2.225&release=2.7.8&remote.application=dubbo-provider&revision=1.0-SNAPSHOT&service.filter=trace&side=consumer&sticky=false&timestamp=1634790913021) failed to connect to server /30.43.56.70:20880 client-side timeout 3000ms (elapsed: 3001ms) from netty client 10.112.2.225 using dubbo version 2.7.8",
//						File:             "NettyClient.java",
//						ApplicationID:    "5",
//						RuntimeID:        "13",
//						ServiceName:      "apm-demo-api",
//						ScopeID:          "fc1f8c074e46a9df505a15c1a94d62cc",
//						CreateTime:       "2021-10-18 11:58:31",
//						UpdateTime:       "2021-10-21 12:36:21",
//					},
//					{
//						Id:               "cd41c7109d96edb62f3bb05380e78ab5",
//						ClassName:        "",
//						Method:           "",
//						Type:             "",
//						EventCount:       1,
//						ExceptionMessage: "",
//						File:             "",
//						ApplicationID:    "5",
//						RuntimeID:        "13",
//						ServiceName:      "apm-demo-api",
//						ScopeID:          "fc1f8c074e46a9df505a15c1a94d62cc",
//						CreateTime:       "",
//						UpdateTime:       "",
//					},
//				},
//			},
//			false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			hub := servicehub.New()
//			events := hub.Events()
//			go func() {
//				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
//			}()
//			err := <-events.Started()
//			if err != nil {
//				t.Error(err)
//				return
//			}
//			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
//			got, err := srv.GetExceptions(tt.args.ctx, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("exceptionService.GetExceptions() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.wantResp) {
//				t.Errorf("exceptionService.GetExceptions() = %v, want %v", got, tt.wantResp)
//			}
//		})
//	}
//}
//
//func Test_exceptionService_GetExceptionEventIds(t *testing.T) {
//	type args struct {
//		ctx context.Context
//		req *pb.GetExceptionEventIdsRequest
//	}
//	tests := []struct {
//		name     string
//		service  string
//		config   string
//		args     args
//		wantResp *pb.GetExceptionEventIdsResponse
//		wantErr  bool
//	}{
//		// TODO: Add test cases.
//		{
//			"case 1",
//			"erda.msp.apm.exception.ExceptionService",
//			`
//		erda.msp.apm.exception.query:
//		`,
//			args{
//				context.TODO(),
//				&pb.GetExceptionEventIdsRequest{
//					// TODO: setup fields
//					ExceptionID: "cd41c7109d96edb62f3bb05380e78ab5",
//					ScopeID:     "fc1f8c074e46a9df505a15c1a94d62cc",
//				},
//			},
//			&pb.GetExceptionEventIdsResponse{
//				// TODO: setup fields.
//				Data: []string{
//					"a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//					"2d812625-4851-4fbd-867d-5b37c42dde9f",
//					"3d6d395b-2edf-4339-a46e-e96127a17a34",
//					"eb9a9c92-ae3d-4b22-8881-3d69f59fdf1a",
//					"162524fb-7b4b-424c-a47e-78b84daba241",
//					"d1db1671-62e3-4215-b39a-1d2d25976be8",
//					"5d8eeee3-3ed0-4446-a7d0-47a4447d1c63",
//					"cbd8794e-ba3f-44a5-a700-1bb3fd91f9d6",
//					"ab17f804-7fd7-480c-ac41-ce2f1336dfa2",
//					"053315a8-8084-45d3-a4a4-1b440f5a79ec",
//					"f3b36c9f-5526-430f-82ff-a4b31b70158c",
//					"c1c6e1e1-faf1-4356-b161-8addaae604ab",
//					"6a9f5680-d481-4ecc-84eb-3f0a50babb73",
//					"fc6ee03f-4876-4369-b537-997751a2df2e",
//					"46b618db-3a5b-46ef-a1d2-29157c899b6a",
//					"e41b2d06-4205-492b-85fe-2045b7bca872",
//					"871f29cf-174b-4203-bb40-1ab1d95fb08e",
//					"a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//				},
//			},
//			false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			hub := servicehub.New()
//			events := hub.Events()
//			go func() {
//				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
//			}()
//			err := <-events.Started()
//			if err != nil {
//				t.Error(err)
//				return
//			}
//			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
//			got, err := srv.GetExceptionEventIds(tt.args.ctx, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("exceptionService.GetExceptionEventIds() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.wantResp) {
//				t.Errorf("exceptionService.GetExceptionEventIds() = %v, want %v", got, tt.wantResp)
//			}
//		})
//	}
//}
//
//func Test_exceptionService_GetExceptionEvent(t *testing.T) {
//	type args struct {
//		ctx context.Context
//		req *pb.GetExceptionEventRequest
//	}
//	tests := []struct {
//		name     string
//		service  string
//		config   string
//		args     args
//		wantResp *pb.GetExceptionEventResponse
//		wantErr  bool
//	}{
//		// TODO: Add test cases.
//		{
//			"case 1",
//			"erda.msp.apm.exception.ExceptionService",
//			`
//		erda.msp.apm.exception:
//		`,
//			args{
//				context.TODO(),
//				&pb.GetExceptionEventRequest{
//					// TODO: setup fields
//					ExceptionEventID: "a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//					ScopeID:          "fc1f8c074e46a9df505a15c1a94d62cc",
//				},
//			},
//			&pb.GetExceptionEventResponse{
//				// TODO: setup fields.
//				Data: &pb.ExceptionEvent{
//					Id:          "a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//					ExceptionID: "cd41c7109d96edb62f3bb05380e78ab5",
//					Metadata: map[string]string{
//						"class":  "org.apache.dubbo.remoting.transport.netty4.NettyClient",
//						"file":   "NettyClient.java",
//						"line":   "174",
//						"method": "doConnect",
//						"type":   "org.apache.dubbo.remoting.RemotingException",
//					},
//					RequestContext: map[string]string{},
//					RequestHeaders: map[string]string{},
//					RequestID:      "",
//					Stacks:         []*pb.Stacks{},
//					Tags: map[string]string{
//						"application_id": "5",
//						"event_id":       "a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//					},
//					Timestamp:      1634790981924,
//					RequestSampled: false,
//				},
//			},
//			false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			//hub := servicehub.New()
//			//events := hub.Events()
//			//go func() {
//			//	hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
//			//}()
//			//err := <-events.Started()
//			//if err != nil {
//			//	t.Error(err)
//			//	return
//			//}
//			//srv := hub.Service(tt.service).(pb.ExceptionServiceServer) GetExceptionEvent
//			var exceptionServiceq *pb.ExceptionServiceServer
//			monkey.PatchInstanceMethod(reflect.TypeOf(exceptionServiceq), "GetExceptionEvent", func(*pb.UnimplementedExceptionServiceServer, context.Context, *pb.GetExceptionEventRequest) (*pb.GetExceptionEventResponse, error) {
//				return &pb.GetExceptionEventResponse{
//					Data: &pb.ExceptionEvent{
//						Id:          "a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//						ExceptionID: "cd41c7109d96edb62f3bb05380e78ab5",
//						Metadata: map[string]string{
//							"class":  "org.apache.dubbo.remoting.transport.netty4.NettyClient",
//							"file":   "NettyClient.java",
//							"line":   "174",
//							"method": "doConnect",
//							"type":   "org.apache.dubbo.remoting.RemotingException",
//						},
//						RequestContext: map[string]string{},
//						RequestHeaders: map[string]string{},
//						RequestID:      "",
//						Stacks:         []*pb.Stacks{},
//						Tags: map[string]string{
//							"application_id": "5",
//							"event_id":       "a9de9ea4-5a4e-4f8a-a61b-94ea85e11089",
//						},
//						Timestamp:      1634790981924,
//						RequestSampled: false,
//					},
//				}, nil
//			})
//
//			s := &exceptionService{
//				p: &provider{
//					Cfg:                nil,
//					Log:                nil,
//					Register:           nil,
//					Cassandra:          nil,
//					exceptionService:   exceptionServiceq,
//					cassandraSession:   nil,
//					ErrorStorageReader: nil,
//					EventStorageReader: nil,
//				},
//			}
//			got, err := s.GetExceptionEvent(tt.args.ctx, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("exceptionService.GetExceptionEvent() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.wantResp) {
//				t.Errorf("exceptionService.GetExceptionEvent() = %v, want %v", got, tt.wantResp)
//			}
//		})
//	}
//}
