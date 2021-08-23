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

package orgFormModal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/home-page-sidebar/i18n"
	"github.com/erda-project/erda/modules/openapi/conf"
)

type OrgFormModal struct {
	ctxBdl protocol.ContextBundle

	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Operations map[string]interface{} `json:"operations"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
}

type Props struct {
	Title  string        `json:"title"`
	Fields []interface{} `json:"fields"`
}

type State struct {
	Visible  bool     `json:"visible"`
	FormData FormData `json:"formData"`
}

type FormData struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Desc        string `json:"desc"`
	IsPublic    bool   `json:"isPublic"`
	Logo        string `json:"logo"`
}

func (o *OrgFormModal) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	o.ctxBdl = bdl
	return nil
}

func (this *OrgFormModal) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	this.State = state
	return nil
}

func (o *OrgFormModal) SetProps() {
	i18nLocale := o.ctxBdl.Bdl.GetLocale(o.ctxBdl.Locale)
	o.Props = Props{
		Title: i18nLocale.Get(i18n.I18nKeyOrgCreate),
		Fields: []interface{}{
			map[string]interface{}{
				"key":       "displayName",
				"label":     i18nLocale.Get(i18n.I18nKeyOrgDisplayName),
				"component": "input",
				"required":  true,
			},
			map[string]interface{}{
				"key":       "name",
				"label":     i18nLocale.Get(i18n.I18nKeyOrgName),
				"component": "input",
				"required":  true,
				"rules": []interface{}{
					map[string]interface{}{
						"msg":     i18nLocale.Get(i18n.I18nKeyOrgCreateRegexp),
						"pattern": "/^(?:[a-z]+|[0-9]+[a-z]+|[0-9]+[-]+[a-z0-9])+(?:(?:(?:[-]*)[a-z0-9]+)+)?$/",
					},
				},
			},
			map[string]interface{}{
				"key":       "desc",
				"label":     i18nLocale.Get(i18n.I18nKeyOrgDesc),
				"component": "textarea",
				"required":  false,
				"componentProps": map[string]interface{}{
					"autoSize": map[string]interface{}{
						"minRows": 4,
						"maxRows": 8,
					},
					"maxLength": 500,
				},
			},
			map[string]interface{}{
				"key":       "isPublic",
				"label":     i18nLocale.Get(i18n.I18nKeyOrgScope),
				"component": "radio",
				"required":  true,
				"componentProps": map[string]interface{}{
					"radioType":   "radio",
					"displayDesc": true,
				},
				"dataSource": map[string]interface{}{
					"static": []interface{}{
						map[string]interface{}{
							"name":  i18nLocale.Get(i18n.I18nKeyOrgScopePrivate),
							"desc":  i18nLocale.Get(i18n.I18nKeyOrgScopePrivateDesc),
							"value": false,
						},
						map[string]interface{}{
							"name":  i18nLocale.Get(i18n.I18nKeyOrgScopePublic),
							"desc":  i18nLocale.Get(i18n.I18nKeyOrgScopePublicDesc),
							"value": true,
						},
					},
				},
			},
			map[string]interface{}{
				"key":       "logo",
				"label":     i18nLocale.Get(i18n.I18nKeyOrgLogo),
				"component": "upload",
				"componentProps": map[string]interface{}{
					"uploadText": i18nLocale.Get(i18n.I18nKeyOrgLogoUpload),
					"sizeLimit":  2,
					"supportFormat": []string{
						"png",
						"jpg",
						"jpeg",
						"gif",
						"bmp",
					},
				},
			},
		},
	}
}

func (o *OrgFormModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := o.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := o.GenComponentState(c); err != nil {
		return err
	}

	o.Name = "orgFormModal"
	o.Type = "FormModal"
	//var visible bool
	//if o.ctxBdl.Identity.OrgID == "" && conf.CreateOrgEnabled() {
	//	visible = true
	//}
	switch event.Operation {
	case apistructs.InitializeOperation:
		o.Operations = map[string]interface{}{
			"submit": map[string]interface{}{
				"key":     "submitOrg",
				"reload":  true,
				"refresh": false,
			},
		}
		o.SetProps()
		o.State.Visible = false
	case apistructs.SubmitOrgOperationKey:
		if !conf.CreateOrgEnabled() {
			return fmt.Errorf("No permission to create organization")
		}
		req := apistructs.OrgCreateRequest{
			Logo:        o.State.FormData.Logo,
			Name:        o.State.FormData.Name,
			DisplayName: o.State.FormData.DisplayName,
			Desc:        o.State.FormData.Desc,
			Admins:      []string{o.ctxBdl.Identity.UserID},
			IsPublic:    o.State.FormData.IsPublic,
			Type:        apistructs.FreeOrgType,
		}
		// personal workbench can only create free org at present
		_, err := o.ctxBdl.Bdl.CreateDopOrg(o.ctxBdl.Identity.UserID, &req)
		if err != nil {
			return err
		}
		o.State.Visible = false
		o.Operations = map[string]interface{}{
			"submit": map[string]interface{}{
				"key":     "submitOrg",
				"reload":  true,
				"refresh": true,
			},
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &OrgFormModal{}
}
