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

package spaceAddButton

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
)

type props struct {
	Text       string      `json:"text"`
	Type       string      `json:"type"`
	Operations interface{} `json:"operations"`
}

type AddButtonCandidateOp struct {
	Click struct {
		Reload  bool                   `json:"reload"`
		Key     string                 `json:"key"`
		Command map[string]interface{} `json:"command"`
	} `json:"click"`
}
type AddButtonCandidate struct {
	Disabled    bool                 `json:"disabled"`
	DisabledTip string               `json:"disabledTip"`
	Key         string               `json:"key"`
	Operations  AddButtonCandidateOp `json:"operations"`
	PrefixIcon  string               `json:"prefixIcon"`
	Text        string               `json:"text"`
}

type ComponentAction struct {
	CtxBdl protocol.ContextBundle
}

type inParams struct {
	ProjectID int64 `json:"projectId"`
}

func (ca *ComponentAction) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalid bundle")
		return err
	}
	ca.CtxBdl = b
	return nil
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := ca.SetBundle(bdl)
	if err != nil {
		return err
	}
	if ca.CtxBdl.InParams == nil {
		return fmt.Errorf("params is empty")
	}

	inParamsBytes, err := json.Marshal(ca.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", ca.CtxBdl.InParams, err)
	}

	var inParams inParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	createAccess, err := ca.CtxBdl.Bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   ca.CtxBdl.Identity.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(inParams.ProjectID),
		Resource: apistructs.TestSpaceResource,
		Action:   apistructs.CreateAction,
	})
	if err != nil {
		return err
	}

	var disabled bool
	var disabledTip string
	i18nLocale := ca.CtxBdl.Bdl.GetLocale(ca.CtxBdl.Locale)
	if !createAccess.Access {
		disabled = true
		disabledTip = i18nLocale.Get(i18n.I18nKeyNoPermission)
	}
	prop := props{
		Text: i18nLocale.Get(i18n.I18nKeyAdd),
		Type: "primary",
	}
	c.Props = prop
	c.Operations = map[string]interface{}{
		"click": struct {
			Reload      bool                   `json:"reload"`
			Key         string                 `json:"key"`
			Command     map[string]interface{} `json:"command"`
			Disabled    bool                   `json:"disabled"`
			DisabledTip string                 `json:"disabledTip"`
		}{
			Reload:      false,
			Key:         "addSpace",
			Disabled:    disabled,
			DisabledTip: disabledTip,
			Command: map[string]interface{}{
				"key":    "set",
				"target": "spaceFormModal",
				"state": map[string]interface{}{
					"visible":  true,
					"formData": nil,
				},
			},
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
