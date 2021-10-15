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

package workbench

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/issue"
)

// personal workbench issue expire days,Not specified, Expired, Due today , Due tomorrow, Due within 7 days, Expires within 30 days, Future: 0
// display undone issue, issues order by priority
var (
	issueUnspecified = "unspecified"
	issueExpired     = "expired"
	issueOneDay      = "oneDay"
	issueTomorrow    = "tomorrow"
	issueSevenDay    = "sevenDay"
	issueThirtyDay   = "thirtyDay"
	issueFeature     = "feature"
	expireDays       = []string{issueUnspecified, issueExpired, issueOneDay, issueTomorrow, issueSevenDay, issueThirtyDay, issueFeature}
	IssuePriorities  = []apistructs.IssuePriority{
		apistructs.IssuePriorityUrgent,
		apistructs.IssuePriorityHigh,
		apistructs.IssuePriorityNormal,
		apistructs.IssuePriorityLow,
	}
	IssueTypes = []apistructs.IssueType{
		apistructs.IssueTypeRequirement,
		apistructs.IssueTypeBug,
		apistructs.IssueTypeTask,
	}
)

type Workbench struct {
	bdl      *bundle.Bundle
	issueSvc *issue.Issue
}

type Option func(*Workbench)

func New(options ...Option) *Workbench {
	is := &Workbench{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithIssue set issue service
func WithIssue(i *issue.Issue) Option {
	return func(w *Workbench) {
		w.issueSvc = i
	}
}

// WithBundle set bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(w *Workbench) {
		w.bdl = bdl
	}
}

// e.workBench.GetUndoneProjectItem concurrent query different expire issue num
func (w *Workbench) SetDiffFinishedIssueNum(req apistructs.IssuePagingRequest, items []*apistructs.WorkbenchProjectItem) error {
	if len(items) == 0 {
		return nil
	}
	var projectIDS []uint64
	for _, item := range items {
		projectIDS = append(projectIDS, item.ProjectDTO.ID)
	}
	nowTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	tomorrow := nowTime.Add(time.Hour * time.Duration(24))
	twoDay := nowTime.Add(time.Hour * time.Duration(24*2))
	sevenDay := nowTime.Add(time.Hour * time.Duration(24*7))
	thirtyDay := nowTime.Add(time.Hour * time.Duration(24*30))
	timeList := [][]int64{
		{0, 0}, // not specified
		{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()},                 // expired
		{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // today expired
		{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // tomorrow expired
		{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // seven day expired
		{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // thirty day expired
		{thirtyDay.Unix(), 0}, //feature expired
	}

	var wg sync.WaitGroup
	var iErr error
	wg.Add(len(expireDays))
	for index, et := range expireDays {
		go func(idx int, ed string) {
			defer wg.Done()
			etIssueReq := apistructs.IssuePagingRequest{}
			etIssueReq.State = req.State
			etIssueReq.StartFinishedAt = timeList[idx][0] * 1000
			if ed == issueUnspecified {
				etIssueReq.IsEmptyPlanFinishedAt = true
			} else {
				if timeList[idx][1] != 0 {
					etIssueReq.EndFinishedAt = timeList[idx][1] * 1000
				}
			}
			etIssueReq.StateBelongs = apistructs.StateBelongs
			etIssueReq.External = true
			etIssueReq.Type = IssueTypes
			etIssueReq.Assignees = req.Assignees
			prosIssueNumList, err := w.issueSvc.GetIssueNumByPros(projectIDS, etIssueReq)
			if err != nil {
				iErr = err
				logrus.Errorf("Failed to get special issue num, request: %v, err: %v", etIssueReq, err)
				return
			}
			issueNumMap := map[uint64]uint64{}
			for _, issueNum := range prosIssueNumList {
				issueNumMap[issueNum.ProjectID] = issueNum.IssueNum
			}
			for _, item := range items {
				if total, existed := issueNumMap[item.ProjectDTO.ID]; existed {
					switch ed {
					case issueUnspecified:
						item.UnSpecialIssueNum = int(total)
					case issueExpired:
						item.ExpiredIssueNum = int(total)
					case issueOneDay:
						item.ExpiredOneDayNum = int(total)
					case issueTomorrow:
						item.ExpiredTomorrowNum = int(total)
					case issueSevenDay:
						item.ExpiredSevenDayNum = int(total)
					case issueThirtyDay:
						item.ExpiredThirtyDayNum = int(total)
					case issueFeature:
						item.FeatureDayNum = int(total)
					}
				}
			}
		}(index, et)
	}
	wg.Wait()
	return iErr
}

func (w *Workbench) GetUndoneProjectItems(req apistructs.WorkbenchRequest, userID string) (*apistructs.WorkbenchResponse, error) {
	req.IssuePagingRequest = apistructs.IssuePagingRequest{
		OrgID:    int64(req.OrgID),
		PageNo:   1,
		PageSize: uint64(req.IssueSize),
		IssueListRequest: apistructs.IssueListRequest{
			StateBelongs: apistructs.UnfinishedStateBelongs,
			Assignees:    []string{userID},
			External:     true,
			OrderBy:      "plan_finished_at asc, FIELD(priority, 'URGENT', 'HIGH', 'NORMAL', 'LOW')",
			Priority:     IssuePriorities,
			Type:         IssueTypes,
			Asc:          true,
		},
	}
	projectMap, err := w.issueSvc.GetIssuesByStates(req)
	if err != nil {
		return nil, err
	}

	res := &apistructs.WorkbenchResponse{}
	res.Data.TotalProject = len(projectMap)
	res.Data.List = make([]*apistructs.WorkbenchProjectItem, 0)
	for _, v := range projectMap {
		res.Data.List = append(res.Data.List, v)
	}
	return res, nil
}
