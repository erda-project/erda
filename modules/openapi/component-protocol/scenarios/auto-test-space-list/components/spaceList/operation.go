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
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/components/spaceFormModal"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

type ListSpaceOperation struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type ClickRowOperation struct {
	Key     string                   `json:"key"`
	Reload  bool                     `json:"reload"`
	Command ClickRowOperationCommand `json:"command"`
}

type ClickRowOperationCommand struct {
	Key     string                 `json:"key"`
	State   map[string]interface{} `json:"state"`
	Target  string                 `json:"target"`
	JumpOut bool                   `json:"jumpOut"`
}

type ClickRowOperationCommandState struct {
}

type dataOperation struct {
	ShowIndex   int                    `json:"showIndex"`
	Key         string                 `json:"key"`
	Reload      bool                   `json:"reload"`
	Text        string                 `json:"text"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip,omitempty"`
	Confirm     string                 `json:"confirm,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Command     map[string]interface{} `json:"command,omitempty"`
	SuccessMsg  string                 `json:"successMsg"`
}

type operationsCommand struct {
	Key    string               `json:"key"`
	State  spaceFormModal.State `json:"state"`
	Target string               `json:"target"`
}

type meta struct {
	ID uint64 `json:"id"`
}

type operationData struct {
	Meta meta `json:"meta"`
}

var (
	edit = dataOperation{
		Key:       "edit",
		Reload:    false,
		Text:      "编辑",
		Command:   map[string]interface{}{},
		Disabled:  false,
		ShowIndex: 1,
	}
	copy = dataOperation{
		Key:       "copy",
		Reload:    true,
		Text:      "复制",
		Confirm:   "是否确认复制",
		Meta:      map[string]interface{}{},
		Disabled:  true,
		ShowIndex: 2,
	}
	export = dataOperation{
		Key:        "export",
		Reload:     true,
		Text:       "导出",
		Confirm:    "是否确认导出",
		Meta:       map[string]interface{}{},
		Disabled:   false,
		SuccessMsg: "导出任务已创建, 请在导入导出记录表中查看进度",
		ShowIndex:  3,
	}
	delete = dataOperation{
		Key:         "delete",
		Reload:      true,
		Text:        "删除",
		Confirm:     "是否确认删除",
		Meta:        map[string]interface{}{},
		DisabledTip: "无法删除",
		Disabled:    true,
		ShowIndex:   4,
	}
	retry = dataOperation{
		Key:       "retry",
		Reload:    true,
		Text:      "重试",
		Meta:      map[string]interface{}{},
		Disabled:  false,
		ShowIndex: 5,
	}
)

func setCommand(space spaceList) map[string]interface{} {
	return map[string]interface{}{
		"key": "set",
		"state": spaceFormModal.State{
			Reload:  false,
			Visible: true,
			FormData: map[string]interface{}{
				"id":   space.ID,
				"name": space.Name,
				"desc": space.Desc,
			},
		},
		"target": "spaceFormModal",
	}
}

func setMeta(space spaceList) map[string]interface{} {
	return map[string]interface{}{
		"id": space.ID,
	}
}

func (a *ComponentSpaceList) handlerListOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	if a.State.PageSize == 0 {
		a.State = state{
			PageSize: DefaultPageSize,
			PageNo:   DefaultPageNo,
		}
	} else if a.State.PageSize != 10 && a.State.PageSize != 20 && a.State.PageSize != 50 && a.State.PageSize != 100 {
		return fmt.Errorf("无效的pageSize")
	}
	spaceList, err := bdl.Bdl.ListTestSpace(inParams.ProjectID, a.State.PageSize, a.State.PageNo)
	if err != nil {
		return err
	}
	a.State.Total = int64(spaceList.Total)

	if err = a.setData(*spaceList); err != nil {
		return err
	}
	return nil
}

func (a *ComponentSpaceList) handlerDeleteOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	cond := operationData{}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &cond); err != nil {
		return err
	}
	if cond.Meta.ID == 0 {
		return fmt.Errorf("无效的 spaceID")
	}
	res, err := bdl.Bdl.GetTestSpace(cond.Meta.ID)
	if err != nil {
		return err
	}
	if res.Status != apistructs.TestSpaceFailed {
		return fmt.Errorf("当前状态不允许删除")
	}

	if err = bdl.Bdl.DeleteTestSpace(cond.Meta.ID, bdl.Identity.UserID); err != nil {
		return err
	}
	if err := a.handlerListOperation(bdl, c, inParams, event); err != nil {
		return err
	}
	return nil
}

func (a *ComponentSpaceList) handlerCopyOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	cond := operationData{}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &cond); err != nil {
		return err
	}
	if cond.Meta.ID == 0 {
		return fmt.Errorf("无效的 spaceID")
	}
	res, err := bdl.Bdl.GetTestSpace(cond.Meta.ID)
	if err != nil {
		return err
	}
	if res.Status != apistructs.TestSpaceOpen {
		return fmt.Errorf("当前状态不允许复制")
	}

	if err = bdl.Bdl.CopyTestSpace(cond.Meta.ID, bdl.Identity.UserID); err != nil {
		return err
	}
	if err := a.handlerListOperation(bdl, c, inParams, event); err != nil {
		return err
	}
	return nil
}

func (a *ComponentSpaceList) handlerExportOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	cond := operationData{}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &cond); err != nil {
		return err
	}
	if cond.Meta.ID == 0 {
		return fmt.Errorf("invalid spaceID")
	}
	if err = bdl.Bdl.ExportTestSpace(bdl.Identity.UserID, apistructs.AutoTestSpaceExportRequest{
		ID:       cond.Meta.ID,
		FileType: apistructs.TestSpaceFileTypeExcel,
		Locale:   bdl.Locale,
	}); err != nil {
		return err
	}
	return nil
}

func (a *ComponentSpaceList) handlerRetryOperation(bdl protocol.ContextBundle, c *apistructs.Component, inParams inParams, event apistructs.ComponentEvent) error {
	cond := operationData{}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &cond); err != nil {
		return err
	}
	if cond.Meta.ID == 0 {
		return fmt.Errorf("无效的 spaceID")
	}
	res, err := bdl.Bdl.GetTestSpace(cond.Meta.ID)
	if err != nil {
		return err
	}
	if res.Status != apistructs.TestSpaceFailed {
		return fmt.Errorf("当前状态不允许重试")
	}
	// 先删除失败的空间
	if err = bdl.Bdl.DeleteTestSpace(cond.Meta.ID, bdl.Identity.UserID); err != nil {
		return err
	}
	// 再复制一个新空间
	if err = bdl.Bdl.CopyTestSpace(*res.SourceSpaceID, bdl.Identity.UserID); err != nil {
		return err
	}
	if err := a.handlerListOperation(bdl, c, inParams, event); err != nil {
		return err
	}
	return nil
}
