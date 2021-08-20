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

package statusviewgroup

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	siteiplist "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-app-site-ip/components/siteIpList"
)

const (
	DefaultSelected = "success"
)

type ComponentViewGroup struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type EdgeAppSiteIPInParam struct {
	ID       int64  `json:"id"`
	AppName  string `json:"appName"`
	SiteName string `json:"siteName"`
}

type EdgeViewGroupState struct {
	Value string `json:"value"`
}

func (c *ComponentViewGroup) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentViewGroup) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentViewGroup) OperationChangeViewGroup() error {
	var (
		vgGroup = EdgeViewGroupState{}
	)

	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &vgGroup)
	if err != nil {
		return err
	}

	if _, ok := c.component.State["value"]; !ok {
		c.component.State["viewGroupSelected"] = DefaultSelected
		c.component.State["value"] = DefaultSelected
	} else {
		c.component.State["viewGroupSelected"] = vgGroup.Value
		c.component.State["value"] = vgGroup.Value
	}

	return nil
}

func (c *ComponentViewGroup) Operation(orgID int64, identity apistructs.Identity) error {
	var (
		err          error
		successTotal int
		failureTotal int
		inParam      = EdgeAppSiteIPInParam{}
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.ctxBundle.InParams)
	if err != nil {
		return fmt.Errorf("marshal id from inparams error: %v", err)
	}

	if err = json.Unmarshal(jsonData, &inParam); err != nil {
		return fmt.Errorf("unmarshal inparam to object error: %v", err)
	}

	res, err := c.ctxBundle.Bdl.GetEdgeInstanceInfo(orgID, inParam.AppName, inParam.SiteName, identity)
	if err != nil {
		return err
	}

	for _, data := range res {
		if siteiplist.GetEdgeApplicationContainerStatus(data.Phase) == "success" {
			successTotal++
		} else {
			failureTotal++
		}
	}

	c.component.Props = getProps(len(res), successTotal, failureTotal, i18nLocale)

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentViewGroup{}
}
