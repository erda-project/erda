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

	"github.com/erda-project/erda/pkg/common/errors"
)

// MSPProjectDB msp_project
type MSPProjectDB struct {
	*gorm.DB
}

func (db *MSPProjectDB) db() *gorm.DB {
	return db.Table(TableMSPProject)
}

func (db *MSPProjectDB) Create(project *MSPProject) (*MSPProject, error) {
	result := db.db().Create(project)

	if result.Error != nil {
		return nil, errors.NewDatabaseError(result.Error)
	}
	value := result.Value.(*MSPProject)
	return value, nil
}

func (db *MSPProjectDB) Update(project *MSPProject) (*MSPProject, error) {
	err := db.db().Model(project).Update(project).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return project, nil
}

func (db *MSPProjectDB) Query(id string) (*MSPProject, error) {
	project := MSPProject{}
	err := db.db().Where("`id` = ?", id).Where("`is_deleted` = ?", false).Find(&project).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &project, err
}

func (db *MSPProjectDB) Delete(id string) (*MSPProject, error) {
	project, err := db.Query(id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	project.IsDeleted = true
	err = db.Model(&project).Update(&project).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return project, err
}
