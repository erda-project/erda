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

package spaceFormModal

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	spec "github.com/erda-project/erda/modules/openapi/component-protocol/component_spec/form_modal"
	"github.com/erda-project/erda/pkg/strutil"
)

type SpaceFormModal struct {
	CtxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      spec.Props             `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
}

type State struct {
	Reload   bool                   `json:"reload"`
	Visible  bool                   `json:"visible"`
	FormData map[string]interface{} `json:"formData"`
}

type operation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type inParams struct {
	ProjectID int64 `json:"projectId"`
}

type AutoTestSpace struct {
	ID        uint64 `json:"id"`
	ProjectID int64  `json:"projectId"`
	Name      string `json:"name"`
	Desc      string `json:"desc"`
}

const (
	regular = "^[a-z\u4e00-\u9fa5A-Z0-9_-]+( )+(.)*$"
)

func (a *SpaceFormModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	a.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	a.Props.Fields = []spec.Field{
		{
			Key:       "name",
			Label:     "空间名",
			Component: "input",
			Required:  true,
			ComponentProps: spec.ComponentProps{
				MaxLength: 50,
			},
			Rules: []spec.FieldRule{
				{
					Pattern: `/^[.a-z\u4e00-\u9fa5A-Z0-9_-\s]*$/`,
					Msg:     "可输入中文、英文、数字、中划线或下划线",
				},
			},
		},
		{
			Key:       "desc",
			Label:     "描述",
			Component: "textarea",
			Required:  false,
			ComponentProps: spec.ComponentProps{
				MaxLength: 1000,
			},
		},
	}
	a.Operations["submit"] = operation{
		Key:    "submit",
		Reload: true,
	}

	if a.CtxBdl.InParams == nil {
		return fmt.Errorf("params is empty")
	}

	inParamsBytes, err := json.Marshal(a.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", a.CtxBdl.InParams, err)
	}

	var inParams inParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}
	// listen on operation
	switch event.Operation {
	case apistructs.AutoTestSpaceSubmitOperationKey:
		if _, ok := a.State.FormData["id"]; ok {
			if err := a.handlerUpdateOperation(a.CtxBdl, c, inParams, event); err != nil {
				return err
			}
		} else {
			if err := a.handlerCreateOperation(a.CtxBdl, c, inParams, event); err != nil {
				return err
			}
		}
		a.State.Reload = true
		a.State.Visible = false
	}
	return nil
}

func (a *SpaceFormModal) handlerCreateOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {

	cond := AutoTestSpace{}
	filterCond, ok := c.State["formData"]
	if ok {
		filterCondS, err := json.Marshal(filterCond)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(filterCondS, &cond); err != nil {
			return err
		}
	}
	if err := strutil.Validate(cond.Desc, strutil.MaxRuneCountValidator(1000)); err != nil {
		return err
	}
	if err := strutil.Validate(cond.Name, strutil.MaxRuneCountValidator(50)); err != nil {
		return err
	}
	reg := regexp.MustCompile("^[a-zA-Z0-9_.|\\-|\\s|\u4e00-\u9fa5]+$")
	if reg == nil { //解释失败，返回nil
		return fmt.Errorf("regexp err")
	}
	if !reg.MatchString(cond.Name) {
		return fmt.Errorf("请输入中文、英文、数字、中划线或下划线")
	}
	err := bdl.Bdl.CreateTestSpace(cond.Name, inParams.ProjectID, cond.Desc, bdl.Identity.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (a *SpaceFormModal) handlerUpdateOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {

	cond := AutoTestSpace{}
	filterCond, ok := c.State["formData"]
	if ok {
		filterCondS, err := json.Marshal(filterCond)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(filterCondS, &cond); err != nil {
			return err
		}
	}
	if err := strutil.Validate(cond.Desc, strutil.MaxRuneCountValidator(1000)); err != nil {
		return err
	}
	if err := strutil.Validate(cond.Name, strutil.MaxRuneCountValidator(50)); err != nil {
		return err
	}
	reg := regexp.MustCompile("^[a-zA-Z0-9_.|\\-|\\s|\u4e00-\u9fa5]+$")
	if reg == nil { //解释失败，返回nil
		return fmt.Errorf("regexp err")
	}
	if !reg.MatchString(cond.Name) {
		return fmt.Errorf("请输入中文、英文、数字、中划线或下划线")
	}
	res, err := bdl.Bdl.GetTestSpace(cond.ID)
	if err != nil {
		return err
	}
	if res.Status != apistructs.TestSpaceOpen {
		return fmt.Errorf("当前状态不允许编辑")
	}
	err = bdl.Bdl.UpdateTestSpace(cond.Name, cond.ID, cond.Desc, bdl.Identity.UserID)
	if err != nil {
		return err
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &SpaceFormModal{
		CtxBdl:     protocol.ContextBundle{},
		Type:       "",
		Operations: map[string]interface{}{},
		State: State{
			Reload: false,
		},
	}
}
