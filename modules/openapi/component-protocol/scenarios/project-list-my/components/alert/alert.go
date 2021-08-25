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

package alert

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	ctxBdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	orgID := ctxBdl.Identity.OrgID
	org, err := ctxBdl.Bdl.GetOrg(orgID)
	if err != nil {
		return err
	}
	if org.BlockoutConfig.BlockDEV || org.BlockoutConfig.BlockProd || org.BlockoutConfig.BlockStage || org.BlockoutConfig.BlockTEST {
		return json.Unmarshal([]byte(`{ "visible": true, "message": "企业处于封网期间，生产环境禁止部署！", "type": "error" }`), &c.Props)
	}
	return json.Unmarshal([]byte(`{ "visible": false }`), &c.Props)
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
