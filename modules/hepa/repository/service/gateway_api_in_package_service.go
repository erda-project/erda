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
	"github.com/xormplus/xorm"

	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GatewayApiInPackageServiceImpl struct {
	engine *orm.OrmEngine
	*SessionHelper
	executor xorm.Interface
}

func NewGatewayApiInPackageServiceImpl() (*GatewayApiInPackageServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayApiInPackageServiceImpl failed")
	}
	return &GatewayApiInPackageServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayApiInPackageServiceImpl) NewSession(helper ...*SessionHelper) (GatewayApiInPackageService, error) {
	var session *SessionHelper
	var err error
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayApiInPackageServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil
	} else {
		session = helper[0]
	}
	return &GatewayApiInPackageServiceImpl{
		engine:        impl.engine,
		executor:      session.session,
		SessionHelper: session,
	}, nil
}

func (impl *GatewayApiInPackageServiceImpl) Update(dao *orm.GatewayApiInPackage) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.executor, dao)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiInPackageServiceImpl) Insert(dao *orm.GatewayApiInPackage) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.executor, dao)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiInPackageServiceImpl) GetByAny(cond *orm.GatewayApiInPackage) (*orm.GatewayApiInPackage, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayApiInPackage{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetByAnyI(impl.executor, bCond, dao)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl *GatewayApiInPackageServiceImpl) SelectByAny(cond *orm.GatewayApiInPackage) ([]orm.GatewayApiInPackage, error) {
	var result []orm.GatewayApiInPackage
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

func (impl *GatewayApiInPackageServiceImpl) Get(id string) (*orm.GatewayApiInPackage, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayApiInPackage{}
	succ, err := orm.Get(impl.executor, dao, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl *GatewayApiInPackageServiceImpl) Delete(packageId, apiId string) error {
	if packageId == "" || apiId == "" {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.executor, &orm.GatewayApiInPackage{}, "package_id = ? and dice_api_id = ?", packageId, apiId)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiInPackageServiceImpl) DeleteByPackageId(packageId string) error {
	if len(packageId) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.executor, &orm.GatewayApiInPackage{}, "package_id = ?", packageId)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayApiInPackageServiceImpl) SelectByApi(id string) ([]orm.GatewayApiInPackage, error) {
	var result []orm.GatewayApiInPackage
	if len(id) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.executor, &result, "dice_api_id = ?", id)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}
