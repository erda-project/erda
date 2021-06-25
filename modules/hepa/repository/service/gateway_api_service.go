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
	"github.com/erda-project/erda/modules/hepa/common"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"

	"github.com/pkg/errors"
	"github.com/xormplus/xorm"
)

type GatewayApiServiceImpl struct {
	engine *orm.OrmEngine
	*SessionHelper
	executor xorm.Interface
}

func NewGatewayApiServiceImpl() (*GatewayApiServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayApiServiceImpl failed")
	}
	return &GatewayApiServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayApiServiceImpl) NewSession(helper ...*SessionHelper) (GatewayApiService, error) {
	var session *SessionHelper
	var err error
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayApiServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil
	} else {
		session = helper[0]
	}
	return &GatewayApiServiceImpl{
		engine:        impl.engine,
		executor:      session.session,
		SessionHelper: session,
	}, nil
}

func (impl *GatewayApiServiceImpl) Insert(api *orm.GatewayApi) error {
	if api == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.executor, api)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiServiceImpl) CountByConsumerId(consumerId string) (int64, error) {
	if len(consumerId) == 0 {
		return 0, errors.New(ERR_INVALID_ARG)
	}
	count, err := orm.Count(impl.executor, &orm.GatewayApi{}, "consumer_id = ?", consumerId)
	if err != nil {
		return 0, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return count, nil
}

func (impl *GatewayApiServiceImpl) GetPageByConsumerId(consumerId string, page *common.Page) (*common.PageQuery, error) {
	if len(consumerId) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	total, err := impl.CountByConsumerId(consumerId)
	if err != nil {
		return nil, errors.Wrap(err, "get total consumer count failed")
	}
	page.SetTotalNum(total)
	if total == 0 {
		return &common.PageQuery{Result: []orm.GatewayApi{}, Page: page}, nil
	}
	var result []orm.GatewayApi
	err = orm.SelectPage(impl.executor.Desc("create_time"), &result, page, "consumer_id = ?", consumerId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return &common.PageQuery{Result: result, Page: page}, nil
}

func (impl *GatewayApiServiceImpl) GetById(id string) (*orm.GatewayApi, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	api := &orm.GatewayApi{}
	succ, err := orm.Get(impl.executor, api, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return api, nil
}

func (impl *GatewayApiServiceImpl) GetPage(options []orm.SelectOption, page *common.Page) (*common.PageQuery, error) {
	total, err := impl.Count(options)
	if err != nil {
		return nil, errors.Wrap(err, "get total count failed")
	}
	page.SetTotalNum(total)
	if total == 0 {
		return &common.PageQuery{Result: []orm.GatewayApi{}, Page: page}, nil
	}
	var result []orm.GatewayApi
	err = orm.SelectPageWithOption(options, impl.executor, &result, page)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return &common.PageQuery{Result: result, Page: page}, nil
}

func (impl *GatewayApiServiceImpl) Count(options []orm.SelectOption) (int64, error) {
	count, err := orm.CountWithOption(options, impl.executor, &orm.GatewayApi{})
	if err != nil {
		return 0, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return count, nil
}

func (impl *GatewayApiServiceImpl) GetByAny(cond *orm.GatewayApi) (*orm.GatewayApi, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	api := &orm.GatewayApi{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetByAnyI(impl.executor, bCond, api)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return api, nil
}

func (impl *GatewayApiServiceImpl) GetRawByAny(cond *orm.GatewayApi) (*orm.GatewayApi, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	api := &orm.GatewayApi{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetRawByAnyI(impl.executor, bCond, api)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return api, nil
}

func (impl *GatewayApiServiceImpl) SelectByGroupId(id string) ([]orm.GatewayApi, error) {
	var result []orm.GatewayApi
	if len(id) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.executor.Desc("create_time"), &result, "group_id = ?", id)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayApiServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.executor, &orm.GatewayApi{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiServiceImpl) RealDeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.RealDelete(impl.executor, &orm.GatewayApi{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiServiceImpl) RealDeleteByRuntimeServiceId(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.RealDelete(impl.executor, &orm.GatewayApi{}, "runtime_service_id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiServiceImpl) Update(api *orm.GatewayApi) error {
	if api == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.executor, api)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiServiceImpl) SelectByOptions(options []orm.SelectOption) ([]orm.GatewayApi, error) {
	var result []orm.GatewayApi
	err := orm.SelectWithOption(options, impl.executor, &result)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayApiServiceImpl) SelectByAny(cond *orm.GatewayApi) ([]orm.GatewayApi, error) {
	var result []orm.GatewayApi
	if cond == nil {
		return result, errors.New(ERR_INVALID_ARG)
	}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	err = orm.SelectByAnyI(impl.executor, bCond, &result)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}
