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

import "github.com/jinzhu/gorm"

// TmcRequestRelationDB .
type TmcRequestRelationDB struct {
	*gorm.DB
}

func (db *TmcRequestRelationDB) GetChildRequestIdsByParentId(parentId string) ([]string, error) {
	var list []TmcRequestRelation
	if err := db.Table(TableRequestRelation).
		Where("`parent_request_id`=? AND is_deleted='N'", parentId).Find(&list).Error; err != nil {
		return nil, err
	}

	var ids []string
	for _, item := range list {
		ids = append(ids, item.ChildRequestId)
	}

	return ids, nil
}

func (db *TmcRequestRelationDB) DeleteRequestRelation(parentId string, childId string) error {
	resp := db.Table(TableRequestRelation).
		Where("`parent_request_id`=? AND `child_request_id`=?", parentId, childId).
		Update("is_deleted", "Y")

	return resp.Error
}
