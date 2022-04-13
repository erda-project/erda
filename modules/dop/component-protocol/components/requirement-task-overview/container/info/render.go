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

package info

import (
	"context"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	cpcommon "github.com/erda-project/erda/modules/dop/component-protocol/components/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/container/simpleChart"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/providers/component-protocol/issueFilter"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "info", func() servicehub.Provider {
		return &Info{}
	})
}

func (i *Info) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := i.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	h := gshelper.NewGSHelper(gs)
	i.Issues = h.GetIssueList()

	conditions := h.GetIssueCondtions()
	issueStateSvc := ctx.Value(types.IssueStateService).(*issuestate.IssueState)
	projectID, err := strconv.ParseUint(cputil.GetInParamByKey(ctx, "projectId").(string), 10, 64)
	if err != nil {
		return err
	}
	stateIDs, err := issueStateSvc.GetIssueStateIDsByTypes(&apistructs.IssueStatesRequest{
		ProjectID:    projectID,
		IssueType:    []apistructs.IssueType{apistructs.IssueTypeTask, apistructs.IssueTypeRequirement},
		StateBelongs: []apistructs.IssueStateBelong{apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
	})
	if err != nil {
		return err
	}
	stats := getStats(i.Issues)

	timeMilli := cpcommon.TimeMilliInDays{
		TodayBegin: cpcommon.MilliFromTime(common.DateTime(time.Now())),
		TodayEnd:   cpcommon.MilliFromTime(common.DateTime(time.Now())) - 1,
		Tomorrow:   cpcommon.MilliFromTime(common.DateTime(time.Now().AddDate(0, 0, 1))),
		SevenDays:  cpcommon.MilliFromTime(common.DateTime(time.Now().AddDate(0, 0, 7))),
		ThirtyDays: cpcommon.MilliFromTime(common.DateTime(time.Now().AddDate(0, 0, 30))),
	}
	i.Data.Data = [][]Data{
		{
			{
				Main: strconv.Itoa(stats.Unclose),
				Sub:  cputil.I18n(ctx, "unfinished"),
				MainLink: buildLink(ConditionsLinkRequest{
					dao.ExpireTypeUnfinished,
					conditions,
					stateIDs,
					timeMilli,
				}),
			},
			{
				Main: strconv.Itoa(stats.Expire),
				Sub:  cputil.I18n(ctx, "expired"),
				MainLink: buildLink(ConditionsLinkRequest{
					dao.ExpireTypeExpired,
					conditions,
					stateIDs,
					timeMilli,
				}),
			},
			{
				Main: strconv.Itoa(stats.Today),
				Sub:  cputil.I18n(ctx, "dueToday"),
				MainLink: buildLink(ConditionsLinkRequest{
					dao.ExpireTypeExpireIn1Day,
					conditions,
					stateIDs,
					timeMilli,
				}),
			},
			{
				Main: strconv.Itoa(stats.Week),
				Sub:  cputil.I18n(ctx, "dueThisWeek"),
				Tip:  cputil.I18n(ctx, "notIncludeDueToday"),
				MainLink: buildLink(ConditionsLinkRequest{
					dao.ExpireTypeExpireIn7Days,
					conditions,
					stateIDs,
					timeMilli,
				}),
			},
		},
		{
			{
				Main: strconv.Itoa(stats.Month),
				Sub:  cputil.I18n(ctx, "dueThisMonth"),
				Tip:  "不包含本日、本周截止数据",
				MainLink: buildLink(ConditionsLinkRequest{
					dao.ExpireTypeExpireIn30Days,
					conditions,
					stateIDs,
					timeMilli,
				}),
			},
			{
				Main: strconv.Itoa(stats.Undefined),
				Sub:  cputil.I18n(ctx, "noDeadlineSpecified"),
				// MainLink: buildLink(dao.ExpireTypeUndefined, conditions),
			},
		},
	}

	return i.SetToProtocolComponent(c)
}

type ConditionsLinkRequest struct {
	ExpiryType dao.ExpireType
	Conditions issueFilter.FrontendConditions
	StateIDs   []int64
	TimeMIlli  cpcommon.TimeMilliInDays
}

func buildLink(req ConditionsLinkRequest) simpleChart.Link {
	conditions := req.Conditions
	switch req.ExpiryType {
	case dao.ExpireTypeExpired:
		conditions.FinishedAtStartEnd = []*int64{&req.TimeMIlli.Tomorrow, nil}
	case dao.ExpireTypeExpireIn1Day:
		conditions.FinishedAtStartEnd = []*int64{&req.TimeMIlli.TodayBegin, &req.TimeMIlli.TodayEnd}
	case dao.ExpireTypeExpireIn7Days:
		conditions.FinishedAtStartEnd = []*int64{&req.TimeMIlli.Tomorrow, &req.TimeMIlli.SevenDays}
	case dao.ExpireTypeExpireIn30Days:
		conditions.FinishedAtStartEnd = []*int64{&req.TimeMIlli.SevenDays, &req.TimeMIlli.ThirtyDays}
	// case dao.ExpireTypeUndefined:
	// 	conditions.FinishedAtStartEnd = []*int64{nil, nil}
	case dao.ExpireTypeUnfinished:
		conditions.States = req.StateIDs
	}
	urlQuery, err := cpcommon.GenerateUrlQueryParams(conditions)
	if err != nil {
		logrus.Errorf("fail to get urlquery, conditions: %v", conditions)
		return simpleChart.Link{}
	}
	return simpleChart.Link{
		Target: common.IssueTarget,
		Params: map[string]interface{}{
			"issueFilter__urlQuery": urlQuery,
		},
	}
}

type Stats struct {
	Unclose   int `json:"unclose,omitempty"`
	Expire    int `json:"expire,omitempty"`
	Today     int `json:"today,omitempty"`
	Week      int `json:"week,omitempty"`
	Month     int `json:"month,omitempty"`
	Undefined int `json:"undefined,omitempty"`
}

func getStats(issues []dao.IssueItem) (s Stats) {
	for _, v := range issues {
		if v.FinishTime != nil {
			continue
		}
		s.Unclose++
		switch v.ExpiryStatus {
		case dao.ExpireTypeExpired:
			s.Expire++
		case dao.ExpireTypeExpireIn1Day:
			s.Today++
		case dao.ExpireTypeExpireIn7Days:
			s.Week++
		case dao.ExpireTypeExpireIn30Days:
			s.Month++
		case dao.ExpireTypeUndefined:
			s.Undefined++
		}
	}
	return
}
