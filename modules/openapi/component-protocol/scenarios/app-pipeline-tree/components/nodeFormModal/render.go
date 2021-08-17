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

package nodeFormModal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/app-pipeline-tree/components/fileTree"
)

var (
	I18nLocalePrefixKey = "wb.content.pipeline.file.tree.node.form."

	addPipeline  = "addPipeline"
	branch       = "branch"
	pipelineName = "pipelineName"
)

func (a *ComponentNodeFormModal) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) (err error) {
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
		return fmt.Errorf("params is emprtt")
	}

	inParamsBytes, err := json.Marshal(a.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", a.CtxBdl.InParams, err)
	}

	var inParams fileTree.InParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	project, err := fileTree.GetOrgIdByProjectId(a.CtxBdl, inParams.ProjectId)
	if err != nil {
		return err
	}

	switch event.Operation {
	case apistructs.FileTreeSubmitOperationKey:
		if err := a.handlerSubmitOperation(bdl, inParams, *project, event); err != nil {
			return err
		}
	}
	i18nLocale := a.CtxBdl.Bdl.GetLocale(a.CtxBdl.Locale)

	a.Props = map[string]interface{}{
		"title": i18nLocale.Get(I18nLocalePrefixKey + addPipeline),
		"fields": []map[string]interface{}{
			{
				"key":       "branch",
				"label":     i18nLocale.Get(I18nLocalePrefixKey + branch),
				"component": "input",
				"required":  true,
				"disabled":  true,
			},
			{
				"key":       "name",
				"label":     i18nLocale.Get(I18nLocalePrefixKey + pipelineName),
				"component": "input",
				"required":  true,
				"componentProps": map[string]interface{}{
					"maxLength": 30,
				},
			},
		},
	}

	return nil
}

func (a *ComponentNodeFormModal) handlerSubmitOperation(ctxBdl protocol.ContextBundle, inParams fileTree.InParams, project apistructs.ProjectDTO, event apistructs.ComponentEvent) error {
	formData := a.State.FormData
	if formData.Branch == "" || formData.Name == "" {
		return fmt.Errorf("name %s or branch %s error: value is empty", formData.Name, formData.Branch)
	}
	if formData.Name == "pipeline.yml" {
		return fmt.Errorf("can not add pipeline.yml yml already exists")
	} else {
		pinode := fmt.Sprintf("%s/%s/tree/%s/.dice/pipelines", inParams.ProjectId, inParams.AppId, formData.Branch)
		var req apistructs.UnifiedFileTreeNodeCreateRequest
		req.Scope = apistructs.FileTreeScopeProjectApp
		req.ScopeID = inParams.ProjectId
		req.Name = formData.Name
		req.UserID = "1"
		req.Type = apistructs.UnifiedFileTreeNodeTypeFile
		req.Pinode = base64.StdEncoding.EncodeToString([]byte(pinode))
		result, err := ctxBdl.Bdl.CreateFileTreeNodes(req, project.OrgID)
		if err != nil {
			return err
		}

		result.Name = formData.Name
		// 数据传递给下一个组件
		a.State.NodeFormModalAddNode = fileTree.NodeFormModalAddNode{
			Results: *result,
			Branch:  formData.Branch,
		}
		a.State.FormData = fileTree.AddNodeOperationCommandStateFormData{}
		a.State.Visible = false
	}
	return nil
}

func (a *ComponentNodeFormModal) marshal(c *apistructs.Component) error {
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
	c.Type = a.Type
	return nil
}

func (a *ComponentNodeFormModal) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state fileTree.AddNodeOperationCommandState
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	var prop map[string]interface{}
	propJson, err := json.Marshal(c.Props)
	if err != nil {
		return err
	}
	err = json.Unmarshal(propJson, &prop)
	if err != nil {
		return err
	}
	a.State = state
	a.Type = c.Type
	a.Props = prop
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentNodeFormModal{}
}
