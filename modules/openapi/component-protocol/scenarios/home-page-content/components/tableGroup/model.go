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

package tableGroup

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type TableGroup struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  struct {
		Visible bool `json:"visible"`
	} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	Data       ProData                `json:"data"`
	State      State                  `json:"state"`
}

type OperationData struct {
	FillMeta string `json:"fillMeta"`
	Meta     Meta   `json:"meta"`
}

type Meta struct {
	PageNo PageNo `json:"pageNo"`
}

type PageNo struct {
	PageNo int `json:"pageNo"`
}

type State struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
	ProsNum  int `json:"prosNum"`
}

type ChangePageNoOperation struct {
	Key      string `json:"key"`
	Reload   bool   `json:"reload"`
	FillMeta string `json:"fillMeta"`
}

type ClickOperation struct {
	Key     string `json:"key"`
	Reload  bool   `json:"reload"`
	Command struct {
		Key     string `json:"key"`
		Target  string `json:"target"`
		JumpOut bool   `json:"jumpOut"`
	} `json:"command"`
}

type TitleProps struct {
	RenderType  string                 `json:"renderType"`
	Value       map[string]interface{} `json:"value"`
	DisplayName string                 `json:"displayName"`
}

type ProItem struct {
	Title struct {
		//IsPureTitle bool `json:"isPureTitle"`
		//PrefixImage string `json:"prefixImage"`
		//Title      string `json:"title"`
		//Level      int    `json:"level"`
		Props      TitleProps             `json:"props"`
		Operations map[string]interface{} `json:"operations"`
	} `json:"title"`
	SubTitle struct {
		Title string `json:"title"`
		Level int    `json:"level"`
		Size  string `json:"size"`
	} `json:"subtitle"`
	Description struct {
		RenderType    string                 `json:"renderType"`
		Visible       bool                   `json:"visible"`
		Value         map[string]interface{} `json:"value"`
		TextStyleName map[string]interface{} `json:"textStyleName"`
	} `json:"description"`
	Table struct {
		Props      map[string]interface{} `json:"props"`
		Data       IssueData              `json:"data"`
		Operations map[string]interface{} `json:"operations"`
	} `json:"table"`
	ExtraInfo ExtraInfo `json:"extraInfo"`
}

type ProData struct {
	List []ProItem `json:"list"`
}

type IssueData struct {
	List []IssueItem `json:"list"`
}

type IssueName struct {
	RenderType  string `json:"renderType"`
	PrefixIcon  string `json:"prefixIcon"`
	Value       string `json:"value"`
	HoverActive string `json:"hoverActive"`
}

type IssueItem struct {
	Id             int64     `json:"id"`
	ProjectId      uint64    `json:"projectId"`
	Type           string    `json:"type"`
	Name           IssueName `json:"name"`
	PlanFinishedAt string    `json:"planFinishedAt"`
	OrgName        string    `json:"orgName"`
}

type ExtraInfo struct {
	Props      ExtraProps             `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

type ExtraProps struct {
	RenderType string `json:"renderType"`
	Value      Value  `json:"value"`
}

type Value struct {
	Text []ValueText `json:"text"`
}

type ValueText struct {
	Text         string `json:"text"`
	OperationKey string `json:"operationKey"`
	Icon         string `json:"icon,omitempty"`
}

type ToSpecificProjectOperation struct {
	Key     string `json:"key"`
	Reload  bool   `json:"reload"`
	Show    bool   `json:"show"`
	Command struct {
		Key     string `json:"key"`
		Target  string `json:"target"`
		JumpOut bool   `json:"jumpOut"`
		State   struct {
			Query struct {
				IssueViewGroupUrlQuery string `json:"issueViewGroup__urlQuery"`
			} `json:"query"`
			Params struct {
				ProjectId string `json:"projectId"`
				OrgName   string `json:"orgName"`
			} `json:"params"`
		} `json:"state"`
		Visible bool `json:"visible"`
	} `json:"command"`
}
