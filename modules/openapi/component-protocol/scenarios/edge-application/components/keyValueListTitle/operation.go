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

package keyvaluelisttitle

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentKVListTitle struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

func (c *ComponentKVListTitle) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

type EdgeKVListStateTitle struct {
	Visible bool `json:"visible,omitempty"`
}

func (c *ComponentKVListTitle) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentKVListTitle) OperationRendering() error {
	var (
		titleState = EdgeKVListStateTitle{}
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return fmt.Errorf("marshal component state error: %v", err)
	}

	err = json.Unmarshal(jsonData, &titleState)
	if err != nil {
		return fmt.Errorf("unmarshal state json data error: %v", err)
	}

	c.component.Props = getProps(titleState.Visible, i18nLocale)

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentKVListTitle{}
}
