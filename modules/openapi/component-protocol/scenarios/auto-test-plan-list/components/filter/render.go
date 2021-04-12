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

package filter

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type AutoTestPlanFilter struct{}

func RenderCreator() protocol.CompRender {
	return &AutoTestPlanFilter{}
}

func (tpm *AutoTestPlanFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if event.Operation.String() == "filter" {
		if _, ok := c.State["values"]; ok {
			fiterDataBytes, err := json.Marshal(c.State["values"])
			if err != nil {
				return err
			}
			var values map[string]string
			if err := json.Unmarshal(fiterDataBytes, &values); err != nil {
				return err
			}
			for k, v := range values {
				c.State[k] = v
			}
		}
	} else {
		c.State["name"] = ""
	}

	return nil
}
