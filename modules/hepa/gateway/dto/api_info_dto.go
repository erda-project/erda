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

package dto

const (
	RT_AUTO_REGISTER = "register"
	RT_AUTO          = "auto"
	RT_MANUAL        = "manual"
	NT_IN            = "inner"
	NT_OUT           = "outer"
	ST_UP            = "asc"
	ST_DOWN          = "desc"
)

type GetApisDto struct {
	From             string
	Method           string
	ConsumerId       string
	RuntimeId        string
	RuntimeServiceId string
	DiceApp          string
	DiceService      string
	ApiPath          string
	RegisterType     string
	NetType          string
	NeedAuth         int
	SortField        string
	SortType         string
	OrgId            string
	ProjectId        string
	Env              string
	Size             int64
	Page             int64
}

type ApiInfoDto struct {
	ApiId string `json:"apiId"`
	// 列表中展示时使用此字段
	Path string `json:"path"`
	// API编辑时用于展现
	DisplayPath string `json:"displayPath"`
	// 若有此字段，API编辑时展现前缀
	DisplayPathPrefix string      `json:"displayPathPrefix,omitempty"`
	OuterNetEnable    bool        `json:"outerNetEnable"`
	RegisterType      string      `json:"registerType"`
	NeedAuth          bool        `json:"needAuth"`
	Method            string      `json:"method,omitempty"`
	Description       string      `json:"description"`
	RedirectAddr      string      `json:"redirectAddr"`
	RedirectPath      string      `json:"redirectPath"`
	RedirectType      string      `json:"redirectType"`
	MonitorPath       string      `json:"monitorPath"`
	Group             GroupDto    `json:"group"`
	CreateAt          string      `json:"createAt"`
	Policies          []PolicyDto `json:"policies"`
	Swagger           interface{} `json:"swagger,omitempty"`
}
