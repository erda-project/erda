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

package dao

import (
	"strings"

	"gorm.io/gorm"
)

type Option func(db *gorm.DB) *gorm.DB

func WhereOption(format string, args ...interface{}) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(format, args...)
	}
}

func MapOption(m map[string]interface{}) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(m)
	}
}

func ByIDOption(id interface{}) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func InOption(col string, values map[interface{}]struct{}) Option {
	var keys []interface{}
	for k := range values {
		keys = append(keys, k)
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(col+" IN (?)", keys)
	}
}

func PageOption(pageSize, pageNo int) Option {
	if pageSize < 0 {
		pageSize = 0
	}
	if pageNo < 1 {
		pageNo = 1
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(pageSize).Offset((pageNo - 1) * pageSize)
	}
}

func OrderByOption(col string, order string) Option {
	if !strings.EqualFold(order, "desc") && !strings.EqualFold(order, "acs") {
		order = "desc"
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(col + " " + strings.ToUpper(order))
	}
}
