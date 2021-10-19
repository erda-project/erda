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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
)

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
	defer monkey.UnpatchAll()
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

	err := p.ProcessEvent(&pb.Content{
		TestPlanID:      1,
		ExecuteTime:     "2006-01-02 15:04:05",
		PassRate:        10.123,
		ApiTotalNum:     100,
	})
	assert.NoError(t, err)
}

func TestParseExecuteTime(t *testing.T) {
	ti := time.Date(2006, 1, 2, 15, 4, 5, 0, time.Local)

	tt := []struct {
		value string
		want  *time.Time
	}{
		{"2006-01-02 15:04:05",
			&ti,
		},
		{"2006-01-02",
			nil,
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, parseExecuteTime(v.value))
	}
}

func TestGetCostTime(t *testing.T) {
	tt := []struct {
		costTimeSec int64
		want        string
	}{
		{
			costTimeSec: 59,
			want:        "00:00:59",
		},

		{
			costTimeSec: 3600,
			want:        "01:00:00",
		},

		{
			costTimeSec: 59*60 + 59,
			want:        "00:59:59",
		},

		{
			costTimeSec: -1,
			want:        "-",
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, getCostTime(v.costTimeSec))
	}
}
