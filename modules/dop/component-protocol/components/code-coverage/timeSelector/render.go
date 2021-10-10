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

package timeSelector

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/code-coverage/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider

	Type       string                 `json:"type,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
	Data       map[string]interface{}
}

type State struct {
	TimeRange []int64 `json:"timeRange"`
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	ca.Type = "DatePicker"
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	oneMonthAgo := now.AddDate(0, 0, -30)
	oneWeekRange := []int64{oneWeekAgo.Unix() * 1000, now.Unix() * 1000}
	oneMonthRange := []int64{oneMonthAgo.Unix() * 1000, now.Unix() * 1000}
	ca.Props = map[string]interface{}{
		"type":       "dateRange",
		"size":       "small",
		"borderTime": true,
		"ranges": map[string]interface{}{
			"一周内": oneWeekRange,
			"一月内": oneMonthRange,
		},
	}
	ca.Operations = map[string]interface{}{
		"onChange": map[string]interface{}{
			"key":    "changeTime",
			"reload": true,
		},
	}
	switch event.Operation {
	case common.DatePickerChangeTimeOperationKey:
		timeRange := event.OperationData["data"].([]int64)
		if timeRange == nil || len(timeRange) < 2 {
			return fmt.Errorf("invalid time range: %v", timeRange)
		}
		ca.State.TimeRange = timeRange
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "timeSelector", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
