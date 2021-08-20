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

package keyvaluelist

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentKVList struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type EdgeKVListState struct {
	AppID   uint64 `json:"appID"`
	Visible bool   `json:"visible,omitempty"`
}

func (c *ComponentKVList) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentKVList) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentKVList) OperationRendering(identity apistructs.Identity) error {
	var (
		kvState = EdgeKVListState{}
	)

	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return fmt.Errorf("marshal component state error: %v", err)
	}

	err = json.Unmarshal(jsonData, &kvState)
	if err != nil {
		return fmt.Errorf("unmarshal state json data error: %v", err)
	}

	if kvState.AppID != 0 {
		edgeApp, err := c.ctxBundle.Bdl.GetEdgeApp(kvState.AppID, identity)
		if err != nil {
			return fmt.Errorf("get edge app error: %v", err)
		}

		resList := make([]EdgeKVListDataItem, 0)

		for key, value := range edgeApp.ExtraData {
			resList = append(resList, EdgeKVListDataItem{
				KeyName: key,
				ValueName: ItemValue{
					RenderType: "copyText",
					Value: ItemValueContent{
						Text:     value,
						CopyText: value,
					},
				},
			})
		}

		c.component.Data = map[string]interface{}{
			"list": resList,
		}
	}

	c.component.Props = getProps(kvState.Visible)

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentKVList{}
}
