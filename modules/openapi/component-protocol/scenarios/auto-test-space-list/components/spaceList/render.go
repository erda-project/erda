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

package spaceList

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	spec "github.com/erda-project/erda/modules/openapi/component-protocol/component_spec/table"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/project-list-my/components/list"
)

type ComponentSpaceList struct {
	CtxBdl protocol.ContextBundle
	State  state                  `json:"state"`
	Props  spec.Props             `json:"props"`
	Data   map[string]interface{} `json:"data"`
}

type state struct {
	Total    int64                   `json:"total"`
	PageSize int64                   `json:"pageSize"`
	PageNo   int64                   `json:"pageNo"`
	Values   common.FilterConditions `json:"values,omitempty"`
}

type props struct {
	RowKey  string    `json:"rowKey,omitempty"`
	Columns []columns `json:"columns,omitempty"`
}

type spaceList struct {
	ID      uint64                 `json:"id"`
	Name    string                 `json:"name"`
	Desc    string                 `json:"desc"`
	Operate dataTask               `json:"operate"`
	Status  map[string]interface{} `json:"status"`
}

type spaceItem struct {
	ID            uint64                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	PrefixImg     string                 `json:"prefixImg"`
	ArchiveStatus ArchiveStatus          `json:"status"`
	ExtraInfos    []ExtraInfos           `json:"extraInfos"`
	Operations    map[string]interface{} `json:"operations"`
}

type ArchiveStatus struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

type ExtraInfos struct {
	Icon    string `json:"icon,omitempty"`
	Text    string `json:"text,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
	Type    string `json:"type,omitempty"`
}

// apistructs.AutoTestSpaceStatus
type columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
}

type inParams struct {
	ProjectID int64 `json:"projectId"`
}

type RenderType string

var (
	RenderTable RenderType = "tableOperation"
)

type dataTask struct {
	RenderType RenderType             `json:"renderType"`
	Value      string                 `json:"value"`
	Operations map[string]interface{} `json:"operations"`
}

func (a *ComponentSpaceList) SetBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil {
		err := fmt.Errorf("invalid bundle")
		return err
	}
	a.CtxBdl = b
	return nil
}

func (a *ComponentSpaceList) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := a.SetBundle(bdl)
	if err != nil {
		return err
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
	switch event.Operation {
	case apistructs.AutoTestSpaceChangePageNoOperationKey, apistructs.AutoTestSpaceChangePageSizeOperationKey, apistructs.InitializeOperation, apistructs.RenderingOperation:
		if err := a.handlerListOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceDeleteOperationKey:
		if err := a.handlerDeleteOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceCopyOperationKey:
		if err := a.handlerCopyOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceRetryOperationKey:
		if err := a.handlerRetryOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	case apistructs.AutoTestSpaceExportOperationKey:
		if err := a.handlerExportOperation(bdl, c, inParams, event); err != nil {
			return err
		}
	}
	c.Operations = getOperations()
	a.Props = getProps()
	return nil
}

func getStatus(req apistructs.AutoTestSpaceStatus) map[string]interface{} {
	res := map[string]interface{}{"renderType": "textWithBadge"}
	if req == apistructs.TestSpaceFailed {
		res["status"] = "error"
		res["value"] = "失败"
	}
	if req == apistructs.TestSpaceCopying {
		res["status"] = "processing"
		res["value"] = "复制中"
	}
	return res
}

func (a *ComponentSpaceList) setData(projectID int64, spaces apistructs.AutoTestSpaceList, statsMap map[uint64]*apistructs.AutoTestSpaceStats) error {
	access, err := a.CtxBdl.Bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   a.CtxBdl.Identity.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectID),
		Resource: apistructs.TestSpaceResource,
		Action:   apistructs.CreateAction,
	})
	if err != nil {
		return err
	}
	i18nLocale := a.CtxBdl.Bdl.GetLocale(a.CtxBdl.Locale)
	lists := []spaceItem{}
	for _, each := range spaces.List {
		var (
			edit = dataOperation{
				Key:         "edit",
				Reload:      false,
				Text:        i18nLocale.Get(i18n.I18nKeyEdit),
				Command:     map[string]interface{}{},
				Disabled:    false,
				ShowIndex:   1,
				DisabledTip: i18nLocale.Get(i18n.I18nKeyNoPermission),
			}
			copyOp = dataOperation{
				Key:         "copy",
				Reload:      true,
				Text:        i18nLocale.Get(i18n.I18nKeyCopy),
				Confirm:     i18nLocale.Get(i18n.I18nKeyCopyConfirm),
				Meta:        map[string]interface{}{},
				Disabled:    true,
				ShowIndex:   2,
				DisabledTip: i18nLocale.Get(i18n.I18nKeyNoPermission),
			}
			export = dataOperation{
				Key:         "export",
				Reload:      true,
				Text:        i18nLocale.Get(i18n.I18nKeyExport),
				Confirm:     i18nLocale.Get(i18n.I18nKeyExportConfirm),
				Meta:        map[string]interface{}{},
				Disabled:    false,
				SuccessMsg:  i18nLocale.Get(i18n.I18nKeyExportSuccessMsg),
				ShowIndex:   3,
				DisabledTip: i18nLocale.Get(i18n.I18nKeyNoPermission),
			}
			deleteOp = dataOperation{
				Key:         "delete",
				Reload:      true,
				Text:        i18nLocale.Get(i18n.I18nKeyDelete),
				Confirm:     i18nLocale.Get(i18n.I18nKeyDeleteConfirm),
				Meta:        map[string]interface{}{},
				DisabledTip: i18nLocale.Get(i18n.I18nKeyDeleteDisabledTip),
				Disabled:    true,
				ShowIndex:   4,
			}
			retry = dataOperation{
				Key:       "retry",
				Reload:    true,
				Text:      i18nLocale.Get(i18n.I18nKeyRetry),
				Meta:      map[string]interface{}{},
				Disabled:  false,
				ShowIndex: 5,
			}
			click = dataOperation{
				Key:    "click",
				Reload: false,
				Command: map[string]interface{}{
					"key":    "goto",
					"target": "project_test_spaceDetail_scenes",
				},
			}
		)
		updatedAt := each.UpdatedAt.Format("2006-01-02 15:04:05")
		text, _ := list.CountActiveTime(updatedAt)
		item := spaceItem{
			ID:          each.ID,
			Title:       each.Name,
			Description: each.Description,
			PrefixImg:   "default_test_case",
			ArchiveStatus: ArchiveStatus{
				Status: each.ArchiveStatus.GetFrontEndStatus(),
				Text:   i18nLocale.Get(fmt.Sprintf("autoTestSpace%s", each.ArchiveStatus)),
			},
			Operations: map[string]interface{}{},
			ExtraInfos: []ExtraInfos{
				{
					Text: fmt.Sprintf("场景集： %v", statsMap[each.ID].SetNum),
				},
				{
					Text: fmt.Sprintf("场景数： %v", statsMap[each.ID].SceneNum),
				},
				{
					Text: fmt.Sprintf("接口数： %v", statsMap[each.ID].StepNum),
				},
				{
					Text:    "更新于 " + text,
					Tooltip: updatedAt,
				},
			},
		}
		if each.Status == apistructs.TestSpaceFailed {
			edit.Command = setCommand(item, each.ArchiveStatus)
			edit.Disabled = true
			item.Operations["a-edit"] = edit
			deleteOp.Meta = setMeta(item)
			deleteOp.Disabled = false
			item.Operations["delete"] = deleteOp
			retry.Meta = setMeta(item)
			export.Meta = setMeta(item)
			export.Disabled = true
			item.Operations["export"] = export
			item.Operations["retry"] = retry
		} else {
			edit.Command = setCommand(item, each.ArchiveStatus)
			copyOp.Meta = setMeta(item)
			deleteOp.Meta = setMeta(item)
			deleteOp.Disabled = true
			copyOp.Disabled = true
			edit.Disabled = true
			export.Disabled = true
			if each.Status == apistructs.TestSpaceOpen && access.Access {
				edit.Disabled = false
				copyOp.Disabled = false
				export.Disabled = false
			}
			item.Operations["a-edit"] = edit
			item.Operations["copy"] = copyOp
			export.Meta = setMeta(item)
			item.Operations["export"] = export
			item.Operations["delete"] = deleteOp
		}

		status := getStatus(each.Status)
		if v, ok := status["value"]; ok {
			item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
				Text: v.(string),
			})
		}
		item.Operations["click"] = click
		lists = append(lists, item)
	}
	a.Data["list"] = lists

	return nil
}

func getOperations() map[string]interface{} {
	return map[string]interface{}{
		"changePageNo": ListSpaceOperation{
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": ListSpaceOperation{
			Key:    "changePageSize",
			Reload: true,
		},
		"clickRow": ClickRowOperation{
			Key:    "clickRow",
			Reload: false,
			Command: ClickRowOperationCommand{
				Key:     "goto",
				State:   map[string]interface{}{},
				Target:  "project_test_spaceDetail_scenes",
				JumpOut: false,
			},
		},
	}
}

func getProps() spec.Props {
	return spec.Props{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		Visible:         true,
	}
}

func RenderCreator() protocol.CompRender {
	return &ComponentSpaceList{
		CtxBdl: protocol.ContextBundle{},
		State:  state{},
		Data:   map[string]interface{}{},
	}
}
