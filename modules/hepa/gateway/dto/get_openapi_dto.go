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

import (
	"strings"

	"github.com/erda-project/erda/modules/hepa/repository/orm"

	"github.com/gin-gonic/gin"
)

type GetOpenapiDto struct {
	DiceArgsDto
	ApiPath string
	Method  string
	Origin  string
}

func NewGetOpenapiDto(c *gin.Context) GetOpenapiDto {
	return GetOpenapiDto{
		DiceArgsDto: NewDiceArgsDto(c),
		ApiPath:     c.Query("apiPath"),
		Method:      c.Query("method"),
		Origin:      c.Query("origin"),
	}
}

func (impl GetOpenapiDto) GenSelectOptions() []orm.SelectOption {
	options := impl.DiceArgsDto.GenSelectOptions()
	if impl.ApiPath != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.FuzzyMatch,
			Column: "api_path",
			Value:  strings.ReplaceAll(impl.ApiPath, `_`, `\_`),
		})
	}
	if impl.Method != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "method",
			Value:  impl.Method,
		})
	}
	if impl.Origin != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "origin",
			Value:  impl.Origin,
		})
	}
	return options
}
