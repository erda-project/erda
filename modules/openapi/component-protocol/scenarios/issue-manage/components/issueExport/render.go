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

package issueExport

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// {
//         "size": "small",
//         "tooltip": "导出",
//         "prefixIcon": "export"
//     }

type IssueExportProps struct {
	Size       string `json:"size"`
	Tooltip    string `json:"tooltip"`
	PrefixIcon string `json:"prefixIcon"`
}

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	c.Props = IssueExportProps{
		Size:       "small",
		Tooltip:    "导出",
		PrefixIcon: "export",
	}
	var click interface{}
	c.Operations = map[string]interface{}{}
	if err := json.Unmarshal([]byte(`{"reload":false,"confirm":"是否确认导出"}`), &click); err != nil {
		return err
	}
	c.Operations["click"] = click
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
