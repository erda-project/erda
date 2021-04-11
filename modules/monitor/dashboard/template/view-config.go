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

package template

import (
	"database/sql/driver"
	"encoding/json"
)

// ViewConfigDTO .
type ViewConfigDTO []*ViewConfigItem

// ViewConfig .
type ViewConfigItem struct {
	W    int64  `json:"w"`
	H    int64  `json:"h"`
	X    int64  `json:"x"`
	Y    int64  `json:"y"`
	I    string `json:"i"`
	View *View  `json:"view"`
}

// Config .
type config struct {
	OptionProps      *map[string]interface{} `json:"optionProps,omitempty"`
	DataSourceConfig interface{}             `json:"dataSourceConfig,omitempty"`
	Option           interface{}             `json:"option,omitempty"`
}

// ViewResp .
type View struct {
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	ChartType      string      `json:"chartType"`
	DataSourceType string      `json:"dataSourceType"`
	StaticData     interface{} `json:"staticData"`
	Config         config      `json:"config"`
	API            *API        `json:"api"`
	Controls       interface{} `json:"controls"`
}

// API .
type API struct {
	URL       string                 `json:"url"`
	Query     map[string]interface{} `json:"query"`
	Body      map[string]interface{} `json:"body"`
	Header    map[string]interface{} `json:"header"`
	ExtraData map[string]interface{} `json:"extraData"`
	Method    string                 `json:"method"`
}

func (vc *ViewConfigDTO) Scan(value interface{}) error {
	if value == nil {
		*vc = ViewConfigDTO{}
		return nil
	}
	t := ViewConfigDTO{}
	if e := json.Unmarshal(value.([]byte), &t); e != nil {
		return e
	}
	*vc = t
	return nil
}

// Value .
func (vc *ViewConfigDTO) Value() (driver.Value, error) {
	if vc == nil {
		return nil, nil
	}
	b, e := json.Marshal(*vc)
	return b, e
}
