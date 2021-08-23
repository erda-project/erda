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

package fileFormModal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-scenes/components/fileTree"
)

func (a *ComponentFileFormModal) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalie bundle")
		return err
	}
	a.CtxBdl = b
	return nil
}

func (a *ComponentFileFormModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = a.SetBundle(bdl)
	if err != nil {
		return err
	}

	err = a.unmarshal(c)
	if err != nil {
		return err
	}

	defer func() {
		fail := a.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	if a.CtxBdl.InParams == nil {
		return fmt.Errorf("params is empty")
	}

	inParamsBytes, err := json.Marshal(a.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", a.CtxBdl.InParams, err)
	}

	var inParams fileTree.InParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	a.Operations = map[string]interface{}{}
	a.Operations["submit"] = Operation{
		Key:    "SubmitScene",
		Reload: true,
	}
	// defaultFields := a.Props.Fields
	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		if err := a.renderHelper(bdl, inParams, event); err != nil {
			return err
		}
	case apistructs.SubmitSceneOperationKey:
		if err := a.renderSubmitHelper(bdl, inParams, event); err != nil {
			return err
		}
	}

	return nil
}

func (a *ComponentFileFormModal) initFields() {
	a.Props = Props{
		Title: "添加",
		Fields: []Entry{
			{
				Key:       "name",
				Label:     "名称",
				Required:  true,
				Component: "input",
				Rules: []Rule{
					{
						Pattern: `/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/`,
						Msg:     "可输入中文、英文、数字、中划线或下划线",
					},
				},
				ComponentProps: ComponentProps{
					MaxLength: 50,
				},
			},
			{
				Key:       "desc",
				Label:     "描述",
				Component: "textarea",
				Required:  false,
				ComponentProps: ComponentProps{
					MaxLength: 1000,
				},
			},
		},
	}
}

func (a *ComponentFileFormModal) initSceneSetFields(bdl protocol.ContextBundle, inParams fileTree.InParams) error {
	a.Props = Props{
		Title: "添加",
		Fields: []Entry{
			{
				Key:       "name",
				Label:     "名称",
				Required:  true,
				Component: "input",
				Rules: []Rule{
					{
						Pattern: `/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/`,
						Msg:     "可输入中文、英文、数字、中划线或下划线",
					},
				},
				ComponentProps: ComponentProps{
					MaxLength: 50,
				},
			},
			{
				Key:       "scenesSet",
				Label:     "选择场景集",
				Component: "select",
				Required:  true,
				ComponentProps: ComponentProps{
					Placeholder: "请选择场景集",
				},
			},
			{
				Key:       "desc",
				Label:     "描述",
				Component: "textarea",
				Required:  false,
				ComponentProps: ComponentProps{
					MaxLength: 1000,
				},
			},
		},
	}

	req := apistructs.SceneSetRequest{
		SpaceID: inParams.SpaceId,
	}
	req.UserID = bdl.Identity.UserID
	rsp, err := bdl.Bdl.GetSceneSets(req)
	if err != nil {
		return err
	}
	for _, v := range rsp {
		if strconv.Itoa(int(v.ID)) == inParams.SetID {
			continue
		}
		a.Props.Fields[1].ComponentProps.Options = append(a.Props.Fields[1].ComponentProps.Options, struct {
			Name  string `json:"name"`
			Value uint64 `json:"value"`
		}{
			v.Name,
			v.ID,
		})
	}
	return nil
}

func (a *ComponentFileFormModal) renderHelper(bdl protocol.ContextBundle, inParams fileTree.InParams, event apistructs.ComponentEvent) error {
	switch a.State.ActionType {
	case "AddScene":
		a.initFields()
		a.Props.Title = "添加场景"
		a.State.FormData = FormData{
			Name:        "",
			Description: "",
		}
	case "addRefSceneSet":
		err := a.initSceneSetFields(bdl, inParams)
		if err != nil {
			return err
		}
		a.Props.Title = "引用场景集"
		a.State.FormData = FormData{
			Name:        "",
			Description: "",
			ScenesSet:   nil,
		}
	case "UpdateSceneSet":
		a.initFields()
		a.Props.Title = "编辑场景集"
		a.Props.Fields = []Entry{a.Props.Fields[0]}
		if err := a.GetSceneSet(bdl); err != nil {
			return err
		}
	case "UpdateScene":
		a.initFields()
		a.Props.Title = "编辑场景"
		if err := a.GetScene(bdl, inParams); err != nil {
			return err
		}
	case "ClickAddSceneSetButton":
		a.initFields()
		a.Props.Title = "添加场景集"
		a.Props.Fields = []Entry{a.Props.Fields[0]}
		a.State.FormData = FormData{
			Name:        "",
			Description: "",
		}
	}
	return nil
}

func (a *ComponentFileFormModal) renderSubmitHelper(bdl protocol.ContextBundle, inParams fileTree.InParams, event apistructs.ComponentEvent) error {
	switch a.State.ActionType {
	case "AddScene":
		if err := a.AddScene(bdl, inParams); err != nil {
			return err
		}
	case "addRefSceneSet":
		if err := a.AddRefSceneSet(bdl, inParams); err != nil {
			return err
		}
	case "UpdateSceneSet":
		if err := a.UpdateSceneSet(bdl, inParams); err != nil {
			return err
		}
	case "UpdateScene":
		if err := a.UpdateScene(bdl, inParams); err != nil {
			return err
		}
	case "ClickAddSceneSetButton":
		if err := a.AddSceneSet(bdl, inParams); err != nil {
			return err
		}
	}
	a.State.Visible = false
	return nil
}

func (a *ComponentFileFormModal) AddSceneSet(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	formData := a.State.FormData
	req := apistructs.SceneSetRequest{
		Name:        formData.Name,
		Description: formData.Description,
		SpaceID:     inParams.SpaceId,
		ProjectId:   inParams.ProjectId,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	_, err := ctxBdl.Bdl.CreateSceneSet(req)
	if err != nil {
		return err
	}
	return nil
}

func (a *ComponentFileFormModal) GetSceneSet(ctxBdl protocol.ContextBundle) error {
	setId := a.State.SceneSetKey
	req := apistructs.SceneSetRequest{
		SetID: uint64(setId),
	}
	req.UserID = a.CtxBdl.Identity.UserID
	s, err := ctxBdl.Bdl.GetSceneSet(req)
	if err != nil {
		return err
	}
	a.State.FormData = FormData{
		Name:        s.Name,
		Description: s.Description,
	}
	return nil
}

func (a *ComponentFileFormModal) GetScene(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	id := a.State.SceneId
	req := apistructs.AutotestSceneRequest{
		SceneID: id,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	s, err := ctxBdl.Bdl.GetAutoTestScene(req)
	if err != nil {
		return err
	}
	a.State.FormData = FormData{
		Name:        s.Name,
		Description: s.Description,
	}
	return nil
}

func (a *ComponentFileFormModal) AddScene(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	formData := a.State.FormData
	setId := a.State.SceneSetKey
	req := apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			SpaceID: inParams.SpaceId,
		},
		Name:        formData.Name,
		Description: formData.Description,
		SetID:       uint64(setId), //formData.SetID,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	_, err := ctxBdl.Bdl.CreateAutoTestScene(req)
	if err != nil {
		return err
	}
	return nil
}

func (a *ComponentFileFormModal) AddRefSceneSet(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	formData := a.State.FormData

	if formData.ScenesSet == nil || *formData.ScenesSet <= 0 {
		return fmt.Errorf("failed to reference scene set, scene set id is empty")
	}

	setId := a.State.SceneSetKey
	req := apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			SpaceID: inParams.SpaceId,
		},
		Name:        formData.Name,
		Description: formData.Description,
		SetID:       uint64(setId), //formData.SetID,
		RefSetID:    *formData.ScenesSet,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	_, err := ctxBdl.Bdl.CreateAutoTestScene(req)
	if err != nil {
		return err
	}
	return nil
}

func (a *ComponentFileFormModal) UpdateSceneSet(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	formData := a.State.FormData
	setId := a.State.SceneSetKey
	req := apistructs.SceneSetRequest{
		Name:        formData.Name,
		Description: formData.Description,
		SpaceID:     inParams.SpaceId,
		SetID:       uint64(setId),
		ProjectId:   inParams.ProjectId,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	_, err := ctxBdl.Bdl.UpdateSceneSet(req)
	if err != nil {
		return err
	}
	return nil
}

func (a *ComponentFileFormModal) UpdateScene(ctxBdl protocol.ContextBundle, inParams fileTree.InParams) error {
	formData := a.State.FormData
	id := a.State.SceneId
	req := apistructs.AutotestSceneSceneUpdateRequest{
		Name:        formData.Name,
		Description: formData.Description,
		SceneID:     id,
	}
	req.UserID = a.CtxBdl.Identity.UserID
	_, err := ctxBdl.Bdl.UpdateAutoTestScene(req)
	if err != nil {
		return err
	}
	return nil
}

func (a *ComponentFileFormModal) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(a.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(a.Props)
	if err != nil {
		return err
	}
	var props interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}

	c.Props = props
	c.State = state
	c.Operations = a.Operations
	// c.Type = a.Type
	return nil
}

func (a *ComponentFileFormModal) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	var prop Props
	propJson, err := json.Marshal(c.Props)
	if err != nil {
		return err
	}
	err = json.Unmarshal(propJson, &prop)
	if err != nil {
		return err
	}
	a.State = state
	// a.Type = c.Type
	a.Props = prop
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFileFormModal{}
}
