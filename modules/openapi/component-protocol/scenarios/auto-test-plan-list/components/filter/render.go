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
