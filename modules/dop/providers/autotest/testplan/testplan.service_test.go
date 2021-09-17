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
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
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

func Test_convertTime(t *testing.T) {
	time, err := convertUTCTime("2020-01-02 04:00:00")
	assert.NoError(t, err)
	s := time.Format("2006-01-02 15:04:05")
	want := "2020-01-01 20:00:00"
	assert.Equal(t, want, s)
}

func Test_processEvent(t *testing.T) {
	svc := &autotestv2.Service{}
	bdl := &bundle.Bundle{}
	p := &TestPlanService{
		bdl:         bdl,
		autoTestSvc: svc,
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(svc), "GetTestPlanV2",
		func(svc *autotestv2.Service, testPlanID uint64, identityInfo apistructs.IdentityInfo) (*apistructs.TestPlanV2, error) {
			return &apistructs.TestPlanV2{
				Name:      "test",
				ProjectID: 1,
			}, nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "QueryNotifiesBySource",
		func(b *bundle.Bundle, orgID string, sourceType, sourceID, itemName, label string, clusterNames ...string) ([]*apistructs.NotifyDetail, error) {
			return []*apistructs.NotifyDetail{
				{
					NotifyGroup: &apistructs.NotifyGroup{},
					NotifyItems: []*apistructs.NotifyItem{
						{},
					},
				},
			}, nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateGroupNotifyEvent",
		func(b *bundle.Bundle, groupNotifyRequest apistructs.EventBoxGroupNotifyRequest) error {
			want := map[string]string{
				"org_name":         "org",
				"project_name":     "project",
				"plan_name":        "test",
				"pass_rate":        "10.12",
				"execute_duration": "10:10:10",
				"api_total_num":    "100",
			}
			assert.True(t, reflect.DeepEqual(want, groupNotifyRequest.Params))
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(b *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				Name:  "project",
				OrgID: uint64(1),
				ID:    uint64(1),
			}, nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg",
		func(b *bundle.Bundle, idOrName interface{}) (*apistructs.OrgDTO, error) {
			return &apistructs.OrgDTO{
				Name: "org",
			}, nil
		})

	err := p.processEvent(&pb.Content{
		TestPlanID:      1,
		ExecuteTime:     "2006-01-02 15:04:05",
		PassRate:        10.123,
		ExecuteDuration: "10:10:10",
		ApiTotalNum:     100,
	})
	assert.NoError(t, err)
}
