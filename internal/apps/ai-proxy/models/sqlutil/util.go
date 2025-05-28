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

package sqlutil

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func HandleOrderBy(sql *gorm.DB, orderBys []string) (*gorm.DB, error) {
	// order by
	if len(orderBys) == 0 {
		sql = sql.Order("updated_at desc")
	} else {
		for _, orderBy := range orderBys {
			// get is desc or asc
			parts := strings.Split(orderBy, " ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid order by: %s", orderBy)
			}
			sql = sql.Order(clause.OrderByColumn{
				Column: clause.Column{Name: parts[0], Raw: false},
				Desc:   strings.EqualFold(parts[1], "desc"),
			})
		}
	}
	return sql, nil
}
