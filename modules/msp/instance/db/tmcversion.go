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

type TmcVersionDB struct {
	*gorm.DB
}

func (db *TmcVersionDB) GetByEngine(engine string, version string) (*TmcVersion, error) {
	if len(engine) <= 0 {
		return nil, nil
	}

	var list []*TmcVersion
	if err := db.Table(TableTmcVersion).
		Where("`engine`=?", engine).
		Where("`version`=?", version).Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}

	return list[0], nil
}

func (db *TmcVersionDB) GetLatestVersionByEngine(engine string) (*TmcVersion, error) {
	if len(engine) <= 0 {
		return nil, nil
	}

	var list []*TmcVersion
	if err := db.Table(TableTmcVersion).
		Where("`engine`=? and version != 'custom'", engine).
		Order("version desc").Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}

	return list[0], nil
}

func (db *TmcVersionDB) UpdateReleaseId(engine string, version string, releaseId string) error {
	return db.Table(TableTmcVersion).
		Where("`engine`=? and `version`=?", engine, version).
		Update("releaseId", releaseId).Error
}
