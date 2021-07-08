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

const EngineMenuJumpKeyPrefix = "MK_JUMP_"

type TmcIniDB struct {
	*gorm.DB
}

func (db TmcIniDB) GetMicroServiceEngineJumpKey(engine string) (string, error) {
	var list []*TmcIni
	if err := db.Table(TableTmcIni).
		Where("`ini_name`=?", EngineMenuJumpKeyPrefix+engine).
		Limit(1).Find(&list).Error; err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", nil
	}

	return list[0].IniValue, nil
}
