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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/requirement-task-overview/common/gshelper"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
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

	stats := getStats(i.Issues)
	i.Data.Data = [][]Data{
		{
			{
				Main: strconv.Itoa(stats.Unclose),
				Sub:  cputil.I18n(ctx, "unfinished"),
			},
			{
				Main: strconv.Itoa(stats.Expire),
				Sub:  cputil.I18n(ctx, "expired"),
			},
			{
				Main: strconv.Itoa(stats.Today),
				Sub:  cputil.I18n(ctx, "dueToday"),
			},
			{
				Main: strconv.Itoa(stats.Week),
				Sub:  cputil.I18n(ctx, "dueThisWeek"),
				Tip:  cputil.I18n(ctx, "notIncludeDueToday"),
			},
		},
		{
			{
				Main: strconv.Itoa(stats.Month),
				Sub:  cputil.I18n(ctx, "dueThisMonth"),
				Tip:  "不包含本日、本周截止数据",
			},
			{
				Main: strconv.Itoa(stats.Undefined),
				Sub:  cputil.I18n(ctx, "noDeadlineSpecified"),
			},
		},
	}

	return i.SetToProtocolComponent(c)
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
