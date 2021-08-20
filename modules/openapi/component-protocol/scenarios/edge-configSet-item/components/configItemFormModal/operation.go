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

package configitemformmodal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-configSet-item/i18n"
	i18r "github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ScopeCommon = "COMMON"
	ScopePublic = "public"
	ScopeSite   = "SITE"
)

type ComponentFormModal struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type ConfigSetUpdate struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}

type ConfigSetCreateCommon struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

type ConfigSetCreateSite struct {
	Key   string  `json:"key"`
	Value string  `json:"value"`
	Scope string  `json:"scope"`
	Sites []int64 `json:"sites"`
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

func (c *ComponentFormModal) OperateRendering(orgID, configSetID int64, identity apistructs.Identity) error {
	var (
		cfgSetState = apistructs.EdgeCfgSetState{}
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return fmt.Errorf("marshal component state error: %v", err)
	}

	err = json.Unmarshal(jsonData, &cfgSetState)
	if err != nil {
		return fmt.Errorf("unmarshal state json data error: %v", err)
	}

	if cfgSetState.FormClear {
		c.component.State["formData"] = map[string]interface{}{}

		cfgSet, err := c.ctxBundle.Bdl.GetEdgeConfigSet(configSetID, identity)
		if err != nil {
			return err
		}

		sites, err := c.ctxBundle.Bdl.ListEdgeSelectSite(orgID, cfgSet.ClusterID, apistructs.EdgeListValueTypeID, identity)
		if err != nil {
			return fmt.Errorf("get avaliable edge clusters error: %v", err)
		}
		c.component.Props = getProps(sites, false, i18nLocale)

	} else if cfgSetState.ConfigSetItemID != 0 {

		cfgSetItem, err := c.ctxBundle.Bdl.GetEdgeCfgSetItem(cfgSetState.ConfigSetItemID, identity)
		if err != nil {
			return fmt.Errorf("get edge cofngiset item error: %v", err)
		}

		formData := map[string]interface{}{
			"id":    cfgSetItem.ID,
			"key":   cfgSetItem.ItemKey,
			"value": cfgSetItem.ItemValue,
			"scope": deConvertScope(cfgSetItem.Scope),
		}

		if cfgSetItem.Scope == convertScope(ScopeSite) {
			formData["sites"] = []string{
				cfgSetItem.SiteName,
			}
		}
		c.component.State["formData"] = formData
		c.component.Props = getProps(nil, true, i18nLocale)
		return nil

	}

	return nil
}

func (c *ComponentFormModal) OperateSubmit(configSetID int64, identity apistructs.Identity) error {
	var (
		updateObject       = ConfigSetUpdate{}
		createPublicObject = ConfigSetCreateCommon{}
		createSiteObject   = ConfigSetCreateSite{}
		itemKey            string
		itemValue          string
		scope              string
		sites              []int64
		formDataJson       []byte
		isUpdate           bool
		isPublic           bool
		err                error
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
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
			if value, ok := formData["scope"].(string); ok {
				if convertScope(value) == ScopePublic {
					isPublic = true
				}
			}
		}
	} else {
		return fmt.Errorf("must provide formdata")
	}

	if isUpdate {
		err = json.Unmarshal(formDataJson, &updateObject)
		if err != nil {
			return fmt.Errorf("unmarshal configset item update form data error: %v", err)
		}

		if err = strutil.Validate(updateObject.Value, strutil.MaxRuneCountValidator(apistructs.EdgeDefaultLagerLength)); err != nil {
			return err
		}

		err = c.ctxBundle.Bdl.UpdateEdgeCfgSetItem(&apistructs.EdgeCfgSetItemUpdateRequest{
			EdgeCfgSetItemCreateRequest: apistructs.EdgeCfgSetItemCreateRequest{
				ItemValue: updateObject.Value,
			},
		}, updateObject.ID, identity)

		if err != nil {
			return fmt.Errorf("update edge configset item error: %v", err)
		}
	} else {
		if isPublic {
			err = json.Unmarshal(formDataJson, &createPublicObject)
			if err != nil {
				return fmt.Errorf("unmarshal common scope type json error: %v", err)
			}
			itemKey = createPublicObject.Key
			itemValue = createPublicObject.Value
			scope = createPublicObject.Scope
		} else {
			err = json.Unmarshal(formDataJson, &createSiteObject)
			if err != nil {
				return fmt.Errorf("unmarshal site scope type json error: %v", err)
			}
			itemKey = createSiteObject.Key
			itemValue = createSiteObject.Value
			scope = createSiteObject.Scope
			sites = createSiteObject.Sites
		}

		if err = validateSubmitData(itemKey, itemValue, i18nLocale); err != nil {
			return err
		}

		req := &apistructs.EdgeCfgSetItemCreateRequest{
			ConfigSetID: configSetID,
			Scope:       convertScope(scope),
			SiteIDs:     sites,
			ItemKey:     itemKey,
			ItemValue:   itemValue,
		}

		err = c.ctxBundle.Bdl.CreateEdgeCfgSetItem(req, identity)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertScope(scope string) string {
	if scope == ScopeCommon {
		return ScopePublic
	} else if scope == ScopeSite {
		return strings.ToLower(ScopeSite)
	}
	return ""
}

func deConvertScope(scope string) string {
	if scope == ScopePublic {
		return ScopeCommon
	} else if scope == strings.ToLower(ScopeSite) {
		return ScopeSite
	}
	return ""
}

func validateSubmitData(itemKey, itemValue string, lr *i18r.LocaleResource) error {
	if err := strutil.Validate(itemKey, strutil.MaxRuneCountValidator(apistructs.EdgeDefaultNameMaxLength)); err != nil {
		return err
	}
	if err := strutil.Validate(itemValue, strutil.MaxRuneCountValidator(apistructs.EdgeDefaultLagerLength)); err != nil {
		return err
	}

	isRight, err := regexp.MatchString(CfgItemKeyMatchPattern, itemKey)
	if err != nil {
		return err
	}

	if !isRight {
		return fmt.Errorf(lr.Get(i18n.I18nKeyInputConfigItemTip))
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFormModal{}
}
