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
	"github.com/gin-gonic/gin"

	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GetOpenLimitRulesDto struct {
	DiceArgsDto
	ConsumerId string
	PackageId  string
}

func NewGetOpenLimitRulesDto(c *gin.Context) GetOpenLimitRulesDto {
	return GetOpenLimitRulesDto{
		DiceArgsDto: NewDiceArgsDto(c),
		ConsumerId:  c.Query("consumerId"),
		PackageId:   c.Query("packageId"),
	}
}

func (impl GetOpenLimitRulesDto) GenSelectOptions() []orm.SelectOption {
	options := impl.DiceArgsDto.GenSelectOptions()
	if impl.ConsumerId != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "consumer_id",
			Value:  impl.ConsumerId,
		})
	}
	if impl.PackageId != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "package_id",
			Value:  impl.PackageId,
		})
	}
	return options
}
