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

package common

import (
	"reflect"
)

type PageQuery struct {
	Result interface{} `json:"result"`
	Page   *Page       `json:"page"`
}

type NewPageQuery struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

func NewPages(list interface{}, total int64) NewPageQuery {
	if list == nil || reflect.ValueOf(list).IsNil() {
		return NewPageQuery{
			List:  []string{},
			Total: 0,
		}
	}
	return NewPageQuery{
		List:  list,
		Total: total,
	}
}

func (query PageQuery) Convert() NewPageQuery {
	return NewPageQuery{
		List:  query.Result,
		Total: query.Page.TotalNum,
	}
}
