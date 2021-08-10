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

type GatewayConsumerServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayConsumerServiceImpl() (*GatewayConsumerServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayConsumerServiceImpl failed")
	}
	return &GatewayConsumerServiceImpl{engine}, nil
}

func (impl *GatewayConsumerServiceImpl) Insert(consumer *orm.GatewayConsumer) error {
	if consumer == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, consumer)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayConsumerServiceImpl) GetDefaultConsumerName(dao *orm.GatewayConsumer) string {
	consumer, err := impl.GetDefaultConsumer(dao)
	if err != nil || consumer == nil {
		return dao.OrgId + "_" + dao.ProjectId + "_" + dao.Env + "_" + dao.Az + "_default"
	}
	return consumer.ConsumerName
}

func (impl *GatewayConsumerServiceImpl) GetById(id string) (*orm.GatewayConsumer, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	consumer := &orm.GatewayConsumer{}
	succ, err := orm.Get(impl.engine, consumer, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return consumer, nil
}

func (impl *GatewayConsumerServiceImpl) Get(cond *orm.GatewayConsumer) (*orm.GatewayConsumer, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	consumer := &orm.GatewayConsumer{}
	succ, err := orm.GetByAny(impl.engine, consumer, cond)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return consumer, nil
}

func (impl *GatewayConsumerServiceImpl) GetDefaultConsumer(cond *orm.GatewayConsumer) (*orm.GatewayConsumer, error) {
	if cond == nil || len(cond.OrgId) == 0 || len(cond.ProjectId) == 0 || len(cond.Env) == 0 || len(cond.Az) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	// backward compatibility
	consumer, err := impl.GetByName(cond.OrgId + "_" + cond.ProjectId + "_" + cond.Env + "_default")
	if err != nil {
		return nil, err
	}
	if consumer != nil && consumer.Az == cond.Az {
		return consumer, nil
	}
	consumer, err = impl.GetByName(cond.OrgId + "_" + cond.ProjectId + "_" + cond.Env + "_" + cond.Az + "_default")
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func (impl *GatewayConsumerServiceImpl) GetByName(name string) (*orm.GatewayConsumer, error) {
	if len(name) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	consumer := &orm.GatewayConsumer{}
	succ, err := orm.Get(impl.engine, consumer, "consumer_name = ?", name)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return consumer, nil
}

func (impl *GatewayConsumerServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayConsumer{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayConsumerServiceImpl) SelectByAny(cond *orm.GatewayConsumer) ([]orm.GatewayConsumer, error) {
	var result []orm.GatewayConsumer
	if cond == nil {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.SelectByAny(impl.engine, &result, cond, "create_time")
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayConsumerServiceImpl) Update(consumer *orm.GatewayConsumer) error {
	if consumer == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, consumer)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayConsumerServiceImpl) CheckUnique(consumer *orm.GatewayConsumer) (bool, error) {
	if consumer == nil {
		return false, errors.New(ERR_INVALID_ARG)
	}
	c := &orm.GatewayConsumer{}
	exist, err := orm.GetByAny(impl.engine, c, &orm.GatewayConsumer{
		OrgId:        consumer.OrgId,
		ProjectId:    consumer.ProjectId,
		Env:          consumer.Env,
		Az:           consumer.Az,
		ConsumerName: consumer.ConsumerName,
	})
	if err != nil {
		return false, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return !exist, nil
}

func (impl *GatewayConsumerServiceImpl) GetPage(options []orm.SelectOption, page *common.Page) (*common.PageQuery, error) {
	total, err := impl.Count(options)
	if err != nil {
		return nil, errors.Wrap(err, "get total count failed")
	}
	page.SetTotalNum(total)
	if total == 0 {
		return &common.PageQuery{Result: []orm.GatewayConsumer{}, Page: page}, nil
	}
	var result []orm.GatewayConsumer
	err = orm.SelectPageWithOption(options, impl.engine.Desc("create_time"), &result, page)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return &common.PageQuery{Result: result, Page: page}, nil
}

func (impl *GatewayConsumerServiceImpl) Count(options []orm.SelectOption) (int64, error) {
	count, err := orm.CountWithOption(options, impl.engine, &orm.GatewayConsumer{})
	if err != nil {
		return 0, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return count, nil
}

func (impl *GatewayConsumerServiceImpl) SelectByOptions(options []orm.SelectOption) ([]orm.GatewayConsumer, error) {
	var result []orm.GatewayConsumer
	err := orm.SelectWithOption(options, impl.engine, &result)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayConsumerServiceImpl) GetByAny(cond *orm.GatewayConsumer) (*orm.GatewayConsumer, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	consumer := &orm.GatewayConsumer{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetByAnyI(impl.engine, bCond, consumer)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return consumer, nil
}
