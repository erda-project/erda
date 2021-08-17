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

type GatewayExtraServiceImpl struct {
	engine *orm.OrmEngine
}

func (impl *GatewayExtraServiceImpl) GetByKeyAndField(key string, field string) (*orm.GatewayExtra, error) {
	if len(key) == 0 || len(field) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	extra := &orm.GatewayExtra{}
	succ, err := orm.Get(impl.engine, extra, "key_id = ? and field = ?", key, field)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return extra, nil
}
