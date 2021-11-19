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

package podsCharts

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	cmpcputil "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var PrimaryColor = []string{"primary8", "primary7", "primary6", "primary5", "primary4", "primary3", "primary2", "primary1"}

func (p *PodsCharts) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if gs == nil {
		return nil
	}
	countValues, ok := (*gs)["countValues"].(map[string]int)
	if !ok {
		logrus.Errorf("invalid count values type: %v", reflect.TypeOf((*gs)["countValues"]))
		return nil
	}
	total := 0
	for _, count := range countValues {
		total += count
	}
	p.Data.Group = nil
	for state, count := range countValues {
		total += count
		p.Data.Group = append(p.Data.Group, p.ParsePodStatus(ctx, state, count, total))
	}
	sort.Slice(p.Data.Group, func(i, j int) bool {
		return p.Data.Group[i][0].Value > p.Data.Group[j][0].Value
	})
	for i := range p.Data.Group {
		color := PrimaryColor[len(PrimaryColor)-1]
		if i < len(PrimaryColor) {
			color = PrimaryColor[i]
		}
		p.Data.Group[i][0].Color = color
	}
	delete(*gs, "countValues")
	p.Transfer(c)
	return nil
}

func (p *PodsCharts) ParsePodStatus(ctx context.Context, state string, cnt, tot int) []Pie {
	color := cmpcputil.PodStatus[state]
	if color == "" {
		color = "Default"
	}
	percent := float64(cnt) / float64(tot) * 100
	status := Pie{
		Name:  cputil.I18n(ctx, state),
		Value: cnt,
		Color: color,
		Total: tot,
		Infos: []Info{
			{
				Main: strconv.FormatInt(int64(cnt), 10),
				Sub:  fmt.Sprintf("%.1f%%", percent),
				Desc: cputil.I18n(ctx, state),
			},
		},
	}
	return []Pie{status}
}

func (p *PodsCharts) Transfer(c *cptype.Component) {
	c.Data = map[string]interface{}{
		"group": p.Data.Group,
	}
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podsCharts", func() servicehub.Provider {
		return &PodsCharts{Type: "LinearDistribution"}
	})
}
