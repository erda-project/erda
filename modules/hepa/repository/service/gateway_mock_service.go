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

type GatewayMockServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayMockServiceImpl() (*GatewayMockServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayMockServiceImpl failed")
	}
	return &GatewayMockServiceImpl{engine}, nil
}

func (impl *GatewayMockServiceImpl) GetMockByAny(cond *orm.GatewayMock) (*orm.GatewayMock, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	find := &orm.GatewayMock{}
	succ, err := orm.GetByAny(impl.engine, find, cond)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return find, nil
}

func (impl *GatewayMockServiceImpl) Update(mock *orm.GatewayMock) error {
	if mock == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, mock)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayMockServiceImpl) Insert(mock *orm.GatewayMock) error {
	if mock == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, mock)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}
