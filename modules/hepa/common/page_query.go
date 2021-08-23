// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
