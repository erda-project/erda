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

type GatewayPolicyServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayPolicyServiceImpl() (*GatewayPolicyServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayPolicyServiceImpl failed")
	}
	return &GatewayPolicyServiceImpl{engine}, nil
}

func (impl *GatewayPolicyServiceImpl) Insert(policy *orm.GatewayPolicy) error {
	if policy == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, policy)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayPolicyServiceImpl) Update(policy *orm.GatewayPolicy) error {
	if policy == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, policy)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayPolicyServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayPolicy{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil

}

func (impl *GatewayPolicyServiceImpl) GetById(id string) (*orm.GatewayPolicy, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	policy := &orm.GatewayPolicy{}
	succ, err := orm.Get(impl.engine, policy, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return policy, nil
}

func (impl *GatewayPolicyServiceImpl) GetByPolicyName(name string, consumerId string) (*orm.GatewayPolicy, error) {
	if len(name) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	policy := &orm.GatewayPolicy{}
	succ, err := orm.Get(impl.engine, policy, "policy_name = ? and consumer_id = ?", name,
		consumerId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return policy, nil
}

func (impl *GatewayPolicyServiceImpl) SelectByCategory(category string) ([]orm.GatewayPolicy, error) {
	var result []orm.GatewayPolicy
	if len(category) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.engine.Desc("create_time"), &result, "category = ?", category)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayPolicyServiceImpl) SelectByCategoryAndConsumer(category string, consumerId string) ([]orm.GatewayPolicy, error) {
	var result []orm.GatewayPolicy
	if len(category) == 0 || len(consumerId) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.engine.Desc("create_time"), &result, "category = ? and consumer_id = ?", category, consumerId)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayPolicyServiceImpl) SelectInIds(ids ...string) ([]orm.GatewayPolicy, error) {
	var result []orm.GatewayPolicy
	if len(ids) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.In(impl.engine, "id", ids).Find(&result)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayPolicyServiceImpl) SelectByAny(cond *orm.GatewayPolicy) ([]orm.GatewayPolicy, error) {
	var result []orm.GatewayPolicy
	if cond == nil {
		return result, errors.New(ERR_INVALID_ARG)
	}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	err = orm.SelectByAnyI(impl.engine, bCond, &result)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayPolicyServiceImpl) GetByAny(cond *orm.GatewayPolicy) (*orm.GatewayPolicy, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	policy := &orm.GatewayPolicy{}
	succ, err := orm.GetByAny(impl.engine, policy, cond)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return policy, nil
}
