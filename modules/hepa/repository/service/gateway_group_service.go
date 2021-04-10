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

	"github.com/erda-project/erda/modules/hepa/common"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GatewayGroupServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayGroupServiceImpl() (*GatewayGroupServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayGroupServiceImpl failed")
	}
	return &GatewayGroupServiceImpl{engine}, nil
}

func (impl *GatewayGroupServiceImpl) Insert(group *orm.GatewayGroup) error {
	if group == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, group)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayGroupServiceImpl) CountByConsumerId(consumerId string) (int64, error) {
	if len(consumerId) == 0 {
		return 0, errors.New(ERR_INVALID_ARG)
	}
	count, err := orm.Count(impl.engine, &orm.GatewayGroup{}, "consumer_id = ?", consumerId)
	if err != nil {
		return 0, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return count, nil
}

func (impl *GatewayGroupServiceImpl) GetPageByConsumerId(consumerId string, page *common.Page) (*common.PageQuery, error) {
	if len(consumerId) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	total, err := impl.CountByConsumerId(consumerId)
	if err != nil {
		return nil, errors.Wrap(err, "get total consumer count failed")
	}
	page.SetTotalNum(total)
	if total == 0 {
		return &common.PageQuery{Result: []orm.GatewayGroup{}, Page: page}, nil
	}
	var result []orm.GatewayGroup
	err = orm.SelectPage(impl.engine.Desc("create_time"), &result, page, "consumer_id = ?", consumerId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return &common.PageQuery{Result: result, Page: page}, nil
}

func (impl *GatewayGroupServiceImpl) GetById(id string) (*orm.GatewayGroup, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	group := &orm.GatewayGroup{}
	succ, err := orm.Get(impl.engine, group, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return group, nil
}

func (impl *GatewayGroupServiceImpl) GetByNameAndConsumerId(name string, id string) (*orm.GatewayGroup, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	group := &orm.GatewayGroup{}
	succ, err := orm.Get(impl.engine, group, "group_name = ? and consumer_id = ?", name, id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return group, nil
}

func (impl *GatewayGroupServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayGroup{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil

}

func (impl *GatewayGroupServiceImpl) Update(group *orm.GatewayGroup) error {
	if group == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, group)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}
