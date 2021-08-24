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
