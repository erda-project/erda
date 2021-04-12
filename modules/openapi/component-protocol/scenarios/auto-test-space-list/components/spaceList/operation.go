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

package spaceList

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	spaceFormModal "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/components/spaceFormModal"
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
	Key         string                 `json:"key"`
	Reload      bool                   `json:"reload"`
	Text        string                 `json:"text"`
	Disabled    bool                   `json:"disabled"`
	DisabledTip string                 `json:"disabledTip,omitempty"`
	Confirm     string                 `json:"confirm,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Command     map[string]interface{} `json:"command,omitempty"`
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
		Key:      "edit",
		Reload:   false,
		Text:     "编辑",
		Command:  map[string]interface{}{},
		Disabled: false,
	}
	copy = dataOperation{
		Key:      "copy",
		Reload:   true,
		Text:     "复制",
		Confirm:  "是否确认复制",
		Meta:     map[string]interface{}{},
		Disabled: true,
	}
	delete = dataOperation{
		Key:         "delete",
		Reload:      true,
		Text:        "删除",
		Confirm:     "是否确认删除",
		Meta:        map[string]interface{}{},
		DisabledTip: "无法删除",
		Disabled:    true,
	}
	retry = dataOperation{
		Key:      "retry",
		Reload:   true,
		Text:     "重试",
		Meta:     map[string]interface{}{},
		Disabled: false,
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
