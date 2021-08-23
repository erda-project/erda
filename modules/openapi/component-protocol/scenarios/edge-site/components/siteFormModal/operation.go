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

package siteformmodal

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	SiteNameLength = 30
)

var (
	SiteNameReservedWordMap = map[string]bool{
		"public": true,
	}
)

type ComponentFormModal struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

func (c *ComponentFormModal) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentFormModal) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentFormModal) OperateRendering(orgID int64, identity apistructs.Identity) error {
	var (
		siteState = apistructs.EdgeSiteState{}
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &siteState)
	if err != nil {
		return err
	}

	if siteState.FormClear {
		c.component.State["formData"] = map[string]interface{}{}
		edgeClusters, err := c.ctxBundle.Bdl.ListEdgeCluster(uint64(orgID), apistructs.EdgeListValueTypeID, identity)
		if err != nil {
			return fmt.Errorf("get avaliable edge clusters error: %v", err)
		}
		c.component.Props = getProps(edgeClusters, false, i18nLocale)
	} else if siteState.SiteID != 0 {
		siteInfo, err := c.ctxBundle.Bdl.GetEdgeSite(siteState.SiteID, identity)
		if err != nil {
			return err
		}
		c.component.Props = getProps(nil, true, i18nLocale)
		c.component.State["formData"] = map[string]interface{}{
			"id":             siteInfo.ID,
			"siteName":       siteInfo.Name,
			"relatedCluster": siteInfo.ClusterName,
			"desc":           siteInfo.Description,
		}
	}

	return nil
}

func (c *ComponentFormModal) OperateSubmit(orgID int64, identity apistructs.Identity) error {
	var (
		err           error
		isUpdate      bool
		formDataJson  []byte
		createRequest SiteFormCreate
		baseRequest   SiteFormSubmitBase
	)

	if data, ok := c.component.State["formData"]; ok {
		formDataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}

		if formData, ok := data.(map[string]interface{}); !ok {
			return fmt.Errorf("request form data assert error")
		} else {
			if _, ok := formData["id"]; ok {
				isUpdate = true
			}
		}
	} else {
		return fmt.Errorf("must provide formdata")
	}

	if isUpdate {
		err = json.Unmarshal(formDataJson, &baseRequest)
		if err != nil {
			return err
		}

		if err = validateSubmitData(baseRequest.SiteName, baseRequest.Desc); err != nil {
			return err
		}

		if err = c.ctxBundle.Bdl.UpdateEdgeSite(
			&apistructs.EdgeSiteUpdateRequest{
				Description: baseRequest.Desc,
			},
			baseRequest.ID,
			identity,
		); err != nil {
			return err
		}
	} else {
		err = json.Unmarshal(formDataJson, &createRequest)
		if err != nil {
			return err
		}

		if err = validateSubmitData(createRequest.SiteName, createRequest.Desc); err != nil {
			return err
		}

		if _, ok := SiteNameReservedWordMap[createRequest.SiteName]; ok {
			return fmt.Errorf("name %s is forbidden, choose a new one please", createRequest.SiteName)
		}

		if err = c.ctxBundle.Bdl.CreateEdgeSite(
			&apistructs.EdgeSiteCreateRequest{
				OrgID:       orgID,
				Name:        createRequest.SiteName,
				ClusterID:   createRequest.RelatedCluster,
				Description: createRequest.Desc,
			},
			identity,
		); err != nil {
			return err
		}
	}

	return nil
}

func validateSubmitData(siteName, desc string) error {
	if err := strutil.Validate(siteName, strutil.MaxRuneCountValidator(SiteNameLength)); err != nil {
		return err
	}

	if err := strutil.Validate(desc, strutil.MaxRuneCountValidator(apistructs.EdgeDefaultLagerLength/2)); err != nil {
		return err
	}

	isRight, err := regexp.MatchString(SiteNameMatchPattern, siteName)
	if err != nil {
		return err
	}
	if !isRight {
		return fmt.Errorf(apistructs.EdgeDefaultRegexpError)
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFormModal{}
}
