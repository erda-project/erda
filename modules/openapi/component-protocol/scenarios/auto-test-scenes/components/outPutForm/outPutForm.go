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

package outPutForm

import (
	"github.com/erda-project/erda/apistructs"
)

func (i *ComponentOutPutForm) SetProps() error {
	paramsNameProp := PropColumn{
		Title: "参数名",
		Key:   PropsKeyParamsName,
		Width: 200,
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
		Width: 200,
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
