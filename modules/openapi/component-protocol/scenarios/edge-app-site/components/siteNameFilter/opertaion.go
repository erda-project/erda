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

package sitenamefilter

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentSiteNameFilter struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

func (c *ComponentSiteNameFilter) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentSiteNameFilter) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentSiteNameFilter) OperationFilter() error {
	var (
		condition apistructs.EdgeSearchCondition
	)

	cdJson, err := json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(cdJson, &condition); err != nil {
		return err
	}

	if condition.Values.Condition != "" {
		c.component.State["searchCondition"] = condition.Values.Condition
	} else {
		c.component.State["searchCondition"] = ""
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentSiteNameFilter{}
}
