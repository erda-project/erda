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
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentInParamsForm struct {
	ctxBdl protocol.ContextBundle

	CommonInParamsForm
}

type CommonInParamsForm struct {
	Version    string                 `json:"version,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	State      State                  `json:"state,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
	Data       Data                   `json:"data,omitempty"`
	InParams   InParams               `json:"inParams,omitempty"`
}

type Data struct {
	List []ParamData `json:"list"`
}

type ParamData struct {
	ParamsName   string                 `json:"name"`
	Desc         string                 `json:"description"`
	DefaultValue string                 `json:"value"`
	Value        string                 `json:"temp"`
	ID           uint64                 `json:"id"`
	Operations   map[string]interface{} `json:"operations"`
}

type InParams struct {
	SceneSetID uint64 `json:"sceneSetId__urlQuery"`
	SpaceId    uint64 `json:"spaceId"`
}

type PropColumn struct {
	Title    string     `json:"title"`
	Key      PropsKey   `json:"key"`
	Width    int64      `json:"width"`
	Flex     int64      `json:"flex"`
	Render   PropRender `json:"render"`
	TitleTip string     `json:"titleTip,omitempty"`
}

type PropRender struct {
	Type             string                 `json:"type,omitempty"`
	ValueConvertType string                 `json:"valueConvertType,omitempty"`
	Options          []PropChangeOption     `json:"options,omitempty"`
	Required         bool                   `json:"required,omitempty"`
	UniqueValue      bool                   `json:"uniqueValue,omitempty"`
	Operations       map[string]interface{} `json:"operations,omitempty"`
	Rules            []PropRenderRule       `json:"rules,omitempty"`
	Props            PropRenderProp         `json:"props,omitempty"`
}

type PropRenderProp struct {
	MaxLength   int64              `json:"maxLength,omitempty"`
	Placeholder string             `json:"placeholder,omitempty"`
	Options     []PropChangeOption `json:"options,omitempty"`
}

type PropRenderRule struct {
	Pattern string `json:"pattern,omitempty"`
	Msg     string `json:"msg,omitempty"`
}

type PropChangeOption struct {
	Label    string             `json:"label"`
	Value    string             `json:"value"`
	IsLeaf   bool               `json:"isLeaf"`
	ToolTip  string             `json:"tooltip"`
	Children []PropChangeOption `json:"children"`
}

func (pct *PropChangeOption) FindValue(v string) *PropChangeOption {
	if pct.Value == v {
		return pct
	}
	for i := range pct.Children {
		k := pct.Children[i].FindValue(v)
		if k != nil {
			return k
		}
	}
	return nil
}

type PropsKey string

const (
	PropsKeyParamsName   PropsKey = "name"
	PropsKeyDesc         PropsKey = "description"
	PropsKeyDefaultValue PropsKey = "value"
	PropsKeyValue        PropsKey = "temp"
)

type OptionValue string

func (ov OptionValue) String() string {
	return string(ov)
}

const (
	BeforeSceneOutPutOptionValue OptionValue = "beforeSceneOutPut" // ??????????????????
	SceneOptionValue             OptionValue = "scene"             // ?????? X
	refSceneSetOptionValue       OptionValue = "refSceneSet"       // ???????????????
	SceneOutPutOptionValue       OptionValue = "sceneOutPut"       // ???????????? X
	MockOptionValue              OptionValue = "mock"              // MOCK
	MockValueOptionValue         OptionValue = "mockValue"         // MOCK???
	GlobalOptionValue            OptionValue = "global"            // ????????????
	GlobalValueOptionValue       OptionValue = "globalValue"       // ??????????????????
	GlobalGlobalOptionValue      OptionValue = "globalGlobal"      // ??????????????????Global
)

type State struct {
	AutotestSceneRequest apistructs.AutotestSceneRequest `json:"autotestSceneRequest"`
	List                 []apistructs.AutoTestSceneInput `json:"list"`
	SceneId              uint64                          `json:"sceneId"`
	SetId                uint64                          `json:"setId"`
}

type OperationBaseInfo struct {
	Key string `json:"key"`
	// ??????????????????
	Text string `json:"text,omitempty"`
	// ????????????
	Confirm string `json:"confirm,omitempty"`
	// ????????????????????????????????????
	Reload      bool   `json:"reload"`
	Disabled    bool   `json:"disabled"`
	DisabledTip string `json:"disabledTip"`

	FillMeta string `json:"fillMeta"`
}

type OpMetaInfo struct {
	apistructs.AutotestSceneRequest
	SelectOption []PropChangeOption `json:"selectOption"`
}

type OperationInfo struct {
	OperationBaseInfo
	Meta OpMetaInfo `json:"meta"`
}

type DeleteOperation OperationInfo
type OnSelectOperation OperationInfo
