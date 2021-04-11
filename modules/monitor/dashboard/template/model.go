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

type templateDTO struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Scope       string         `json:"scope"`
	ScopeID     string         `json:"scopeId"`
	ViewConfig  *ViewConfigDTO `json:"viewConfig"`
	CreatedAt   int64          `json:"createdAt"`
	UpdatedAt   int64          `json:"updatedAt"`
	Version     string         `json:"version"`
	Type        int64          `json:"type"`
}

type templateUpdate struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	ViewConfig  *ViewConfigDTO `json:"viewConfig"`
}

type templateOverview struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	ScopeID     string `json:"scopeId"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Version     string `json:"version"`
	Type        int64  `json:"type"`
}

type templateResp struct {
	TemplateDTO []*templateOverview `json:"list"`
	Total       int                 `json:"total"`
}

type templateSearch struct {
	ID       string `query:"id"`
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int64  `query:"pageNo" validate:"gte=1" default:"20"`
	PageSize int64  `query:"pageSize" validate:"gte=1" default:"20"`
	Type     int64  `query:"type"`
	Name     string `query:"name"`
}

type templateType struct {
	Type int64
}
