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

package scene_and_api_filter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/test-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKeyTestDashboard, "scene_and_api_filter", func() servicehub.Provider {
		return &Filter{}
	})
}

func (f *Filter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)

	if err := f.initFromProtocol(ctx, c); err != nil {
		return err
	}
	times := f.State.Values.Time
	if len(times) != 2 {
		times = []int64{time.Now().AddDate(0, 0, -7).Unix() * 1000, time.Now().Unix() * 1000}
		f.State.Values.Time = times
	}
	timeStart := time.Unix(times[0]/1000, 0).Format("2006-01-02 15:04:05")
	timeEnd := time.Unix(times[1]/1000, 0).Format("2006-01-02 15:04:05")

	fmt.Println(timeStart)
	fmt.Println(timeEnd)

	h.SetAtSceneAndApiTimeFilter(gshelper.AtSceneAndApiTimeFilter{
		TimeStart: timeStart, TimeEnd: timeEnd,
	})

	if err := f.setState(); err != nil {
		return err
	}

	return f.setToComponent(c)
}

func (f *Filter) setState() error {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	customProps := CustomProps{
		AllowClear: false,
		Ranges: Ranges{
			Week:  []int64{weekAgo.Unix() * 1000, now.Unix() * 1000},
			Month: []int64{monthAgo.Unix() * 1000, now.Unix() * 1000},
		},
		SelectableTime: f.State.Values.Time,
	}

	b, err := json.Marshal(&customProps)
	if err != nil {
		return err
	}
	customPropsMap := make(map[string]interface{}, 0)
	if err = json.Unmarshal(b, &customPropsMap); err != nil {
		return err
	}

	if len(f.State.Conditions) == 1 {
		f.State.Conditions[0].CustomProps = customPropsMap
	} else if len(f.State.Conditions) == 0 {
		f.State.Conditions = make([]filter.PropCondition, 0, 1)
		f.State.Conditions[0].CustomProps = customPropsMap
	}

	return nil
}
