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

package testplan

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
)

func Test_TestPlanService_UpdateTestPlanV2ByHook(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.TestPlanUpdateByHookRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.TestPlanUpdateByHookResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dop.autotest.testplan.TestPlanService",
		//			`
		//erda.core.dop.autotest.testplan:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.TestPlanUpdateByHookRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.TestPlanUpdateByHookResponse{
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
			srv := hub.Service(tt.service).(pb.TestPlanServiceServer)
			got, err := srv.UpdateTestPlanByHook(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestPlanService.UpdateTestPlanV2ByHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("TestPlanService.UpdateTestPlanV2ByHook() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
