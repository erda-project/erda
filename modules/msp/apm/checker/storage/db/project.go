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

package db

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// ProjectDB .
type ProjectDB struct {
	*gorm.DB
}

func (db *ProjectDB) query() *gorm.DB {
	return db.Table(TableProject).Where("`is_deleted`=?", "N")
}

func (db *ProjectDB) GetByFields(fields map[string]interface{}) (*Project, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), projectFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*Project
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *ProjectDB) GetByID(id int64) (*Project, error) {
	return db.GetByFields(map[string]interface{}{
		"ID": id,
	})
}

func (db *ProjectDB) GetByProjectID(projectID int64) (*Project, error) {
	return db.GetByFields(map[string]interface{}{
		"ProjectID": projectID,
	})
}
