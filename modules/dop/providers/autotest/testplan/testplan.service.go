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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	pb "github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/autotest/testplan/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	autotestv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
)

type TestPlanService struct {
	p   *provider
	db  db.TestPlanDB
	bdl *bundle.Bundle

	autoTestSvc *autotestv2.Service
}

func (s *TestPlanService) WithAutoTestSvc(sv *autotestv2.Service) {
	s.autoTestSvc = sv
}

func (s *TestPlanService) UpdateTestPlanByHook(ctx context.Context, req *pb.TestPlanUpdateByHookRequest) (*pb.TestPlanUpdateByHookResponse, error) {
	logrus.Info("start testplan execute callback")
	if req.Content.TestPlanID == 0 {
		return nil, apierrors.ErrUpdateTestPlan.MissingParameter("testPlanID")
	}
	go func() {
		err := s.processEvent(req.Content)
		if err != nil {
			logrus.Errorf("failed to processEvent, err: %s", err.Error())
		}
	}()

	fields := make(map[string]interface{}, 0)
	fields["pass_rate"] = req.Content.PassRate
	fields["execute_time"] = req.Content.ExecuteTime
	fields["execute_api_num"] = req.Content.ApiTotalNum
	if err := s.db.UpdateTestPlanV2(req.Content.TestPlanID, fields); err != nil {
		return nil, err
	}
	return &pb.TestPlanUpdateByHookResponse{Data: req.Content.TestPlanID}, nil
}

func (s *TestPlanService) processEvent(req *pb.Content) error {
	eventName := "autotest-plan-execute"

	testPlan, err := s.autoTestSvc.GetTestPlanV2(req.TestPlanID, apistructs.IdentityInfo{
		InternalClient: "dop",
	})
	if err != nil {
		return err
	}
	project, err := s.bdl.GetProject(testPlan.ProjectID)
	if err != nil {
		return err
	}
	org, err := s.bdl.GetOrg(project.OrgID)
	if err != nil {
		return err
	}
	orgID := strconv.FormatUint(project.OrgID, 10)
	projectID := strconv.FormatUint(project.ID, 10)
	notifyDetails, err := s.bdl.QueryNotifiesBySource(orgID, "project", projectID, eventName, "")
	for _, notifyDetail := range notifyDetails {
		if notifyDetail.NotifyGroup == nil {
			continue
		}
		notifyItem := notifyDetail.NotifyItems[0]
		params := map[string]string{
			"org_name":         org.Name,
			"project_name":     project.Name,
			"plan_name":        testPlan.Name,
			"pass_rate":        fmt.Sprintf("%.2f", req.PassRate),
			"execute_duration": req.ExecuteDuration,
			"api_total_num":    fmt.Sprintf("%d", req.ApiTotalNum),
		}
		marshal, _ := json.Marshal(params)
		logrus.Debugf("testplan params :%s", string(marshal))

		eventboxReqContent := apistructs.GroupNotifyContent{
			SourceName:            "",
			SourceType:            "project",
			SourceID:              projectID,
			NotifyName:            notifyDetail.Name,
			NotifyItemDisplayName: notifyItem.DisplayName,
			Channels:              []apistructs.GroupNotifyChannel{},
			Label:                 notifyItem.Label,
			CalledShowNumber:      notifyItem.CalledShowNumber,
			OrgID:                 int64(project.OrgID),
		}

		err := s.bdl.CreateGroupNotifyEvent(apistructs.EventBoxGroupNotifyRequest{
			Sender:        "adapter",
			GroupID:       notifyDetail.NotifyGroup.ID,
			Channels:      notifyDetail.Channels,
			NotifyItem:    notifyItem,
			NotifyContent: &eventboxReqContent,
			Params:        params,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func convertUTCTime(tm string) (time.Time, error) {
	executeTime, err := time.Parse("2006-01-02 15:04:05", tm)
	if err != nil {
		return time.Time{}, err
	}
	// executeTime is UTC
	m, err := time.ParseDuration("-8h")
	if err != nil {
		return time.Time{}, err
	}
	return executeTime.Add(m), nil
}
