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

package service

import (
	"github.com/pkg/errors"

	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type IniServiceImpl struct {
	engine *orm.OrmEngine
}

func NewIniServiceImpl() (*IniServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new IniServiceImpl failed")
	}
	return &IniServiceImpl{engine}, nil
}

func (impl *IniServiceImpl) GetValueByName(name string) (string, error) {
	if len(name) == 0 {
		return "", errors.New(ERR_INVALID_ARG)
	}
	ini := &orm.Ini{}
	succ, err := orm.Get(impl.engine, ini, "ini_name = ?", name)
	if err != nil {
		return "", errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return "", nil
	}
	return ini.IniValue, nil

}
