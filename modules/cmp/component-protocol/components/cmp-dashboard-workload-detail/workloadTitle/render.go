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

package workloadTitle

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "workloadTitle", func() servicehub.Provider {
		return &ComponentWorkloadTitle{}
	})
}

func (t *ComponentWorkloadTitle) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	workloadID := t.State.WorkloadID
	splits := strings.Split(workloadID, "_")
	if len(splits) != 3 {
		return fmt.Errorf("invalid workload id: %s", workloadID)
	}
	kind, name := splits[0], splits[2]
	typ := ""
	switch kind {
	case string(apistructs.K8SDeployment):
		typ = "Deployment"
	case string(apistructs.K8SReplicaSet):
		typ = "ReplicaSet"
	case string(apistructs.K8SDaemonSet):
		typ = "DaemonSet"
	case string(apistructs.K8SStatefulSet):
		typ = "StatefulSet"
	case string(apistructs.K8SJob):
		typ = "Job"
	case string(apistructs.K8SCronJob):
		typ = "CronJob"
	}

	t.Props.Title = fmt.Sprintf("%s: %s", typ, name)
	return nil
}
