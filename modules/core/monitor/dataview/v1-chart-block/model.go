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

package block

import (
	"database/sql/driver"
	"encoding/json"
	"net/url"
)

// DashboardBlockDTO .
type DashboardBlockDTO struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Desc       string         `json:"desc"`
	Scope      string         `json:"scope"`
	ScopeID    string         `json:"scopeId"`
	ViewConfig *ViewConfigDTO `json:"viewConfig"`
	DataConfig *dataConfigDTO `json:"-"`
	CreatedAt  int64          `json:"createdAt"`
	UpdatedAt  int64          `json:"updatedAt"`
	Version    string         `json:"version"`
}

func (dash *DashboardBlockDTO) ReplaceVCWithDynamicParameter(query url.Values) *DashboardBlockDTO {
	dash.ViewConfig.replaceWithQuery(query)
	return dash
}

// SystemBlockUpdate .
type SystemBlockUpdate struct {
	Name       *string        `json:"name"`
	Desc       *string        `json:"desc"`
	ViewConfig *ViewConfigDTO `json:"viewConfig"`
}

// UserBlockUpdate .
type UserBlockUpdate struct {
	Name       *string        `json:"name"`
	Desc       *string        `json:"desc"`
	ViewConfig *ViewConfigDTO `json:"viewConfig"`
	DataConfig *dataConfigDTO `json:"dataConfig"`
}

// DashboardBlockResp .
type dashboardBlockResp struct {
	DashboardBlocks []*DashboardBlockDTO `json:"list"`
	Total           int                  `json:"total"`
}

// DataConfigDTO .
type dataConfigDTO []dataConfig

// DataConfig .
type dataConfig struct {
	I          string      `json:"i"`
	StaticData interface{} `json:"staticData"`
}

// Scan .
func (ls *dataConfigDTO) Scan(value interface{}) error {
	if value == nil {
		*ls = dataConfigDTO{}
		return nil
	}
	t := dataConfigDTO{}
	if e := json.Unmarshal(value.([]byte), &t); e != nil {
		return e
	}
	*ls = t
	return nil
}

// Value .
func (ls *dataConfigDTO) Value() (driver.Value, error) {
	if ls == nil {
		return nil, nil
	}
	b, e := json.Marshal(*ls)
	return b, e
}
