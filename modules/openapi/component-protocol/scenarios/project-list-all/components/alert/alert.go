// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
