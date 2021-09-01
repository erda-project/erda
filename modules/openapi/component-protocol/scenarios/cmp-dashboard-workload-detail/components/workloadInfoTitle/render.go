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

package workloadInfoTitle

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &ComponentWorkloadInfoTitle{}
}

func (t *ComponentWorkloadInfoTitle) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	t.Props.Title = "Workload Detail"
	t.Props.Size = "small"
	return nil
}
