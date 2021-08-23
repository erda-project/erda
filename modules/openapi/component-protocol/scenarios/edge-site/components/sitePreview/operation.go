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

package sitepreview

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	edgesite "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-site/i18n"
)

type InfoData struct {
	Info map[string]interface{} `json:"info"`
}

type PropsRender struct {
	Type      string                 `json:"type"`
	DataIndex string                 `json:"dataIndex"`
	Props     map[string]interface{} `json:"props,omitempty"`
}

type PreviewState struct {
	SiteID int64 `json:"siteID"`
}

type ComponentSitePreview struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

func (c *ComponentSitePreview) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentSitePreview) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentSitePreview) OperationRendering(identity apistructs.Identity) error {

	var (
		siteState = PreviewState{}
		siteName  string
		shell     string
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return fmt.Errorf("marshal component state error: %v", err)
	}

	if err := json.Unmarshal(jsonData, &siteState); err != nil {
		return err
	}

	if siteState.SiteID == 0 {
		return nil
	}

	res, err := c.ctxBundle.Bdl.GetEdgeSiteInitShell(siteState.SiteID, identity)
	if err != nil {
		return fmt.Errorf("render site init shell error: %v", err)
	}

	for key, data := range res {
		siteName = key
		for _, value := range data.([]interface{}) {
			shell = shell + value.(string) + "\n"
		}
		break
	}

	c.component.Data = edgesite.StructToMap(InfoData{
		Info: map[string]interface{}{
			"siteName":      siteName,
			"firstStep":     i18nLocale.Get(i18n.I18nKeyAddNodeTip),
			"secondStep":    i18nLocale.Get(i18n.I18nKeyAddNodeCommandTip),
			"operationCode": shell,
		},
	})

	c.component.Props = getProps(i18nLocale)
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentSitePreview{}
}
