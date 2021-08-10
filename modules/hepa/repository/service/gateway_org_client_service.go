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

type GatewayOrgClientServiceImpl struct {
	engine *orm.OrmEngine
	*SessionHelper
	executor xorm.Interface
}

func NewGatewayOrgClientServiceImpl() (*GatewayOrgClientServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayOrgClientServiceImpl failed")
	}
	return &GatewayOrgClientServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayOrgClientServiceImpl) NewSession(helper ...*SessionHelper) (GatewayOrgClientService, error) {
	var session *SessionHelper
	var err error
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayOrgClientServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil

	} else {
		session = helper[0]
	}
	return &GatewayOrgClientServiceImpl{
		engine:        impl.engine,
		executor:      session.session,
		SessionHelper: session,
	}, nil
}

func (impl *GatewayOrgClientServiceImpl) Update(orgClient *orm.GatewayOrgClient) error {
	if orgClient == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.executor, orgClient)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayOrgClientServiceImpl) Insert(orgClient *orm.GatewayOrgClient) error {
	if orgClient == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.executor, orgClient)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayOrgClientServiceImpl) GetByAny(cond *orm.GatewayOrgClient) (*orm.GatewayOrgClient, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	orgClient := &orm.GatewayOrgClient{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetByAnyI(impl.executor, bCond, orgClient)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return orgClient, nil
}

func (impl *GatewayOrgClientServiceImpl) SelectByAny(cond *orm.GatewayOrgClient) ([]orm.GatewayOrgClient, error) {
	var result []orm.GatewayOrgClient
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

func (impl *GatewayOrgClientServiceImpl) GetById(id string) (*orm.GatewayOrgClient, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	orgClient := &orm.GatewayOrgClient{}
	succ, err := orm.Get(impl.executor, orgClient, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return orgClient, nil
}

func (impl *GatewayOrgClientServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.executor, &orm.GatewayOrgClient{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayOrgClientServiceImpl) CheckUnique(orgClient *orm.GatewayOrgClient) (bool, error) {
	if orgClient == nil {
		return false, errors.New(ERR_INVALID_ARG)
	}
	c := &orm.GatewayOrgClient{}
	exist, err := orm.GetByAny(impl.engine, c, &orm.GatewayOrgClient{
		OrgId: orgClient.OrgId,
		Name:  orgClient.Name,
	})
	if err != nil {
		return false, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return !exist, nil
}
