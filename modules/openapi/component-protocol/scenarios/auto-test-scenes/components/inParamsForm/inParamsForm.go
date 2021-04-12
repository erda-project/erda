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

package inParamsForm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
)

func (i *ComponentInParamsForm) SetProps() {
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
	defaultValueProp := PropColumn{
		Title:    "引用值",
		TitleTip: "在执行计划中生效",
		Key:      PropsKeyDefaultValue,
		Flex:     2,
		Render: PropRender{
			Type:             "inputSelect",
			ValueConvertType: "last",
			Required:         true,
			Props: PropRenderProp{
				Placeholder: "可选择表达式",
				Options: []PropChangeOption{
					{
						Label:  "前置场景出参",
						Value:  BeforeSceneOutPutOptionValue.String(),
						IsLeaf: false,
					},
					{
						Label:  "mock",
						Value:  MockOptionValue.String(),
						IsLeaf: false,
					},
					{
						Label:  "全局变量入参",
						Value:  GlobalOptionValue.String(),
						IsLeaf: false,
					},
				},
			},
			Operations: make(map[string]interface{}),
		},
	}
	o := OperationInfo{}
	o.Key = apistructs.AutoTestSceneInputOnSelectOperationKey.String()
	o.Reload = true
	o.FillMeta = "selectOption"
	defaultValueProp.Render.Operations[apistructs.AutoTestSceneInputOnSelectOperationKey.String()] = o
	valueProp := PropColumn{
		Title:    "调试值",
		TitleTip: "仅在当前场景执行时生效",
		Key:      PropsKeyValue,
		Width:    100,
	}
	i.Props["temp"] = []PropColumn{paramsNameProp, descProp, defaultValueProp, valueProp}
}

func (i *ComponentInParamsForm) RenderListInParamsForm() error {
	rsp, err := i.ctxBdl.Bdl.ListAutoTestSceneInput(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}
	list := make([]ParamData, 0, 0)
	for _, v := range rsp {
		pd := ParamData{
			ParamsName:   v.Name,
			Desc:         v.Description,
			DefaultValue: v.Value,
			Value:        v.Temp,
			ID:           v.ID,
		}
		list = append(list, pd)
	}

	i.Data.List = list

	i.Operations = make(map[string]interface{})
	o := apistructs.Operation{}
	o.Key = apistructs.AutoTestSceneInputUpdateOperationKey.String()
	o.Reload = true
	i.Operations[apistructs.AutoTestSceneInputUpdateOperationKey.String()] = o
	return nil
}

func (i *ComponentInParamsForm) RenderUpdateInParamsForm() error {
	req := apistructs.AutotestSceneInputUpdateRequest{
		AutotestSceneRequest: i.State.AutotestSceneRequest,
		List:                 i.State.List,
	}
	req.UserID = i.ctxBdl.Identity.UserID
	_, err := i.ctxBdl.Bdl.UpdateAutoTestSceneInputs(req)
	if err != nil {
		return err
	}
	return nil
}

func (i *ComponentInParamsForm) RenderOnSelect(opsData interface{}) error {
	req, err := GetOpsInfo(opsData)
	if err != nil {
		return err
	}
	selectOptions := req.SelectOption
	var idIndex int
	pcl := i.Props["temp"].([]PropColumn)
	for id, p := range pcl {
		if p.Key == PropsKeyDefaultValue {
			idIndex = id
			break
		}
	}
	// 每一个已经选择的下拉菜单
	for _, value := range selectOptions {
		// 搜索对应value值的option
		for index := range pcl[idIndex].Render.Props.Options {
			op := (&pcl[idIndex].Render.Props.Options[index]).FindValue(value.Value)
			if op == nil {
				continue
			}
			children := make([]PropChangeOption, 0, 0)
			if strings.HasPrefix(value.Value, BeforeSceneOutPutOptionValue.String()) {
				// 前置场景出参
				_, list, err := i.ctxBdl.Bdl.ListAutoTestScene(i.State.AutotestSceneRequest)
				if err != nil {
					return err
				}
				for _, v := range list {
					if v.ID == i.State.AutotestSceneRequest.SceneID {
						break
					}
					o := PropChangeOption{
						Label:  v.Name,
						Value:  SceneOptionValue.String() + strconv.FormatInt(int64(v.ID), 10),
						IsLeaf: false,
					}
					children = append(children, o)
				}
			} else if strings.HasPrefix(value.Value, SceneOptionValue.String()) {
				// 场景
				// 从场景的value中获取id的string
				idStr := strings.TrimPrefix(op.Value, SceneOptionValue.String())
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return err
				}
				req := i.State.AutotestSceneRequest
				req.SceneID = uint64(id)
				dbScene, err := i.ctxBdl.Bdl.GetAutoTestScene(req)
				if err != nil {
					return err
				}

				// if the scene references the scene set
				if dbScene.RefSetID > 0 {
					// list sceneSet to children
					var autotestSceneRequest apistructs.AutotestSceneRequest
					autotestSceneRequest.UserID = i.ctxBdl.Identity.UserID
					autotestSceneRequest.SetID = dbScene.RefSetID
					autotestSceneRequest.SpaceID = i.InParams.SpaceId
					_, list, err := i.ctxBdl.Bdl.ListAutoTestScene(autotestSceneRequest)
					if err != nil {
						return err
					}
					for _, v := range list {
						if v.ID == i.State.AutotestSceneRequest.SceneID {
							break
						}
						o := PropChangeOption{
							Label:  v.Name,
							Value:  refSceneSetOptionValue.String() + strconv.FormatInt(int64(v.ID), 10) + "_" + strconv.FormatInt(int64(dbScene.ID), 10),
							IsLeaf: false,
						}
						children = append(children, o)
					}
				} else {
					// list scene output to children
					list, err := i.ctxBdl.Bdl.ListAutoTestSceneOutput(req)
					if err != nil {
						return err
					}
					for _, v := range list {
						str, err := i.GetSceneOutPutValue(dbScene.ID, v.Name)
						if err != nil {
							return err
						}
						o := PropChangeOption{
							Label:  v.Name,
							Value:  str,
							IsLeaf: true,
						}
						children = append(children, o)
					}
				}
			} else if strings.HasPrefix(value.Value, refSceneSetOptionValue.String()) {
				// get sceneID and ref sceneID
				sceneRefSetValue := strings.TrimPrefix(op.Value, refSceneSetOptionValue.String())
				sceneRefSet := strings.Split(sceneRefSetValue, "_")
				if len(sceneRefSet) != 2 {
					return fmt.Errorf("error find ref sceneSet output")
				}
				refSceneID, err := strconv.ParseInt(sceneRefSet[0], 10, 64)
				if err != nil {
					return err
				}
				sceneID, err := strconv.ParseUint(sceneRefSet[1], 10, 64)
				if err != nil {
					return err
				}

				// list scene output to children
				req := i.State.AutotestSceneRequest
				req.SceneID = uint64(refSceneID)
				dbScene, err := i.ctxBdl.Bdl.GetAutoTestScene(req)
				if err != nil {
					return err
				}
				list, err := i.ctxBdl.Bdl.ListAutoTestSceneOutput(req)
				if err != nil {
					return err
				}
				for _, v := range list {
					str, err := i.GetSceneOutPutValue(sceneID, strconv.FormatInt(int64(dbScene.ID), 10)+"_"+v.Name)
					if err != nil {
						return err
					}
					o := PropChangeOption{
						Label:  v.Name,
						Value:  str,
						IsLeaf: true,
					}
					children = append(children, o)
				}
			} else if strings.HasPrefix(value.Value, MockOptionValue.String()) {
				for _, v := range MockString {
					o := PropChangeOption{
						Label:  v,
						Value:  "${{ random." + v + " }}",
						IsLeaf: true,
					}
					children = append(children, o)
				}
			} else if strings.HasPrefix(value.Value, GlobalOptionValue.String()) {
				cfgReq := apistructs.AutoTestGlobalConfigListRequest{Scope: "project-autotest-testcase", ScopeID: strconv.Itoa(int(i.ctxBdl.InParams["projectId"].(float64)))}
				cfgReq.UserID = i.ctxBdl.Identity.UserID
				cfgs, err := i.ctxBdl.Bdl.ListAutoTestGlobalConfig(cfgReq)
				if err != nil {
					return err
				}
				cfgChildren0 := make([]PropChangeOption, 0, 0)
				for _, cfg := range cfgs {
					// Header 是自动带上去的
					// cfgChildren1, cfgChildren2, cfgChildren3 := make([]Input, 0, 0), make([]Input, 0, 0), make([]Input, 0, 0)
					cfgChildren1, cfgChildren3 := make([]PropChangeOption, 0, 0), make([]PropChangeOption, 0, 0)
					// for k := range cfg.APIConfig.Header {
					// 	cfgChildren2 = append(cfgChildren2, Input{Label: k, Value: "{{" + k + "}}", IsLeaf: true})
					// }
					for _, v := range cfg.APIConfig.Global {
						cfgChildren3 = append(cfgChildren3, PropChangeOption{Label: v.Name, Value: "${{ configs.autotest." + v.Name + " }}", IsLeaf: true})
					}
					// cfgChildren1 = append(cfgChildren1, Input{Label: "Header", Value: "Header", IsLeaf: false, Children: cfgChildren2})
					cfgChildren1 = append(cfgChildren1, PropChangeOption{Label: "Global", Value: "Global", IsLeaf: false, Children: cfgChildren3})
					cfgChildren0 = append(cfgChildren0, PropChangeOption{Label: cfg.DisplayName, Value: cfg.DisplayName, IsLeaf: false, Children: cfgChildren1})
				}
				children = cfgChildren0
			}
			op.Children = children
			break
		}
	}
	i.Props["temp"] = pcl
	return nil
}

func (i *ComponentInParamsForm) GetSceneOutPutValue(sceneName uint64, name string) (string, error) {
	valStr := fmt.Sprintf("${{ outputs.%v.%s }}", sceneName, name)
	return valStr, nil
}

func (i *ComponentInParamsForm) Filter() []ParamData {
	var (
		list []ParamData //新添加且未保存的行不在数据库里，但是需要返回
	)
	for _, v := range i.State.List {
		data := ParamData{
			ParamsName:   v.Name,
			Desc:         v.Description,
			DefaultValue: v.Value,
			Value:        v.Temp,
		}
		list = append(list, data)
	}
	return list
}
