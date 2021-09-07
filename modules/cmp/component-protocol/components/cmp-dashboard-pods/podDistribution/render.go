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

package PodDistribution

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (pd *PodDistribution) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if gs == nil {
		return nil
	}
	countValues, ok := (*gs)["countValues"].(map[string]int)
	if !ok {
		logrus.Errorf("invalid count values type: %v", reflect.TypeOf((*gs)["countValues"]))
		return nil
	}
	total := 0
	pd.Data.Lists = nil
	for state, count := range countValues {
		total += count
		pd.Data.Lists = append(pd.Data.Lists, pd.ParsePodStatus(ctx, state, count))
	}
	pd.Data.Total = total
	sort.Slice(pd.Data.Lists, func(i, j int) bool {
		return pd.Data.Lists[i].Label < pd.Data.Lists[j].Label
	})
	return nil
}

func (pd *PodDistribution) ParsePodStatus(ctx context.Context, state string, cnt int) List {
	status := List{
		Tip:   fmt.Sprintf("%s %d/%d", cputil.I18n(ctx, state), cnt, pd.Data.Total),
		Value: cnt,
		Label: fmt.Sprintf("%s %d", cputil.I18n(ctx, state), cnt),
	}
	switch state {
	case "Completed":
		status.Color = "steelBlue"
	case "ContainerCreating":
		status.Color = "orange"
	case "CrashLoopBackOff":
		status.Color = "red"
	case "Error":
		status.Color = "maroon"
	case "Evicted":
		status.Color = "darkgoldenrod"
	case "ImagePullBackOff":
		status.Color = "darksalmon"
	case "Pending":
		status.Color = "teal"
	case "Running":
		status.Color = "lightgreen"
	case "Terminating":
		status.Color = "brown"
	}
	return status
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podDistribution", func() servicehub.Provider {
		return &PodDistribution{Type: "LinearDistribution"}
	})
}
