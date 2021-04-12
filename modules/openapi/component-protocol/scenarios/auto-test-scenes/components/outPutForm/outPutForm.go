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

package outPutForm

import (
	"github.com/erda-project/erda/apistructs"
)

func (i *ComponentOutPutForm) SetProps() error {
	paramsNameProp := PropColumn{
		Title: "参数名",
		Key:   PropsKeyParamsName,
		Width: 100,
		Render: PropRender{
			Type:        "input",
			Required:    true,
			UniqueValue: true,
			Rules: []PropRenderRule{
				{
					Pattern: "/^[a-zA-Z0-9_-]*$/",
					Msg:     "参数名为英文、数字、中划线或下划线",
				},
			},
			Props: PropRenderProp{MaxLength: 50},
		},
	}
	descProp := PropColumn{
		Title: "描述",
		Key:   PropsKeyDesc,
		Width: 140,
		Render: PropRender{
			Type:  "input",
			Props: PropRenderProp{MaxLength: 1000},
		},
	}
	ValueProp := PropColumn{
		Title: "值",
		Key:   PropsKeyValue,
		Flex:  2,
		Render: PropRender{
			Type:     "select",
			Required: true,
			Props:    PropRenderProp{},
		},
	}
	lt, err := i.RenderOnChange()
	if err != nil {
		return err
	}
	ValueProp.Render.Props.Options = lt
	i.Props["temp"] = []PropColumn{paramsNameProp, descProp, ValueProp}
	return nil
}

func (i *ComponentOutPutForm) RenderListOutPutForm() error {
	rsp, err := i.ctxBdl.Bdl.ListAutoTestSceneOutput(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}
	list := []ParamData{}
	for _, v := range rsp {
		pd := ParamData{
			Name:        v.Name,
			Description: v.Description,
			Value:       v.Value,
			ID:          v.ID,
		}
		list = append(list, pd)
	}
	i.Data.List = list
	if err = i.SetProps(); err != nil {
		return err
	}
	return nil
}

func (i *ComponentOutPutForm) RenderUpdateOutPutForm() error {
	req := apistructs.AutotestSceneOutputUpdateRequest{
		AutotestSceneRequest: i.State.AutotestSceneRequest,
		List:                 i.State.List,
	}
	req.UserID = i.ctxBdl.Identity.UserID

	_, err := i.ctxBdl.Bdl.UpdateAutoTestSceneOutput(req)
	if err != nil {
		return err
	}
	return nil
}

// 可编辑器的初始值
func (i *ComponentOutPutForm) RenderOnChange() ([]PropChangeOption, error) {
	list, err := i.ctxBdl.Bdl.ListAutoTestSceneStep(i.State.AutotestSceneRequest)
	if err != nil {
		return nil, err
	}

	req := apistructs.AutotestListStepOutPutRequest{
		List: list,
	}
	req.UserID = i.ctxBdl.Identity.UserID
	mp, err := i.ctxBdl.Bdl.ListAutoTestStepOutput(req)

	var lt []PropChangeOption
	for k, v := range mp {
		lt = append(lt, PropChangeOption{
			Label: k,
			Value: v,
		})
	}
	return lt, nil
}
