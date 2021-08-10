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

type GatewayRuntimeServiceServiceImpl struct {
	engine *orm.OrmEngine
	*SessionHelper
	executor xorm.Interface
}

func NewGatewayRuntimeServiceServiceImpl() (*GatewayRuntimeServiceServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayRuntimeServiceServiceImpl failed")
	}
	return &GatewayRuntimeServiceServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayRuntimeServiceServiceImpl) NewSession(helper ...*SessionHelper) (GatewayRuntimeServiceService, error) {
	var session *SessionHelper
	var err error
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayRuntimeServiceServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil
	} else {
		session = helper[0]
	}
	return &GatewayRuntimeServiceServiceImpl{
		engine:        impl.engine,
		executor:      session.session,
		SessionHelper: session,
	}, nil
}

func (impl *GatewayRuntimeServiceServiceImpl) Update(dao *orm.GatewayRuntimeService) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.executor, dao)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayRuntimeServiceServiceImpl) Insert(dao *orm.GatewayRuntimeService) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.executor, dao)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayRuntimeServiceServiceImpl) GetByAny(cond *orm.GatewayRuntimeService) (*orm.GatewayRuntimeService, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayRuntimeService{}
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

func (impl *GatewayRuntimeServiceServiceImpl) SelectByAny(cond *orm.GatewayRuntimeService) ([]orm.GatewayRuntimeService, error) {
	var result []orm.GatewayRuntimeService
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

func (impl *GatewayRuntimeServiceServiceImpl) Get(id string) (*orm.GatewayRuntimeService, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayRuntimeService{}
	succ, err := orm.Get(impl.executor, dao, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl *GatewayRuntimeServiceServiceImpl) Delete(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.executor, &orm.GatewayRuntimeService{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayRuntimeServiceServiceImpl) CreateIfNotExist(session *xorm.Session, dao *orm.GatewayRuntimeService) (*orm.GatewayRuntimeService, error) {
	if session == nil || dao.RuntimeName == "" || dao.AppId == "" || dao.ServiceName == "" ||
		dao.ProjectId == "" || dao.Workspace == "" || dao.ClusterName == "" {
		return nil, errors.Errorf("invalid dao:%+v", dao)
	}
	exist := &orm.GatewayRuntimeService{}
	// check exist without lock
	succ, err := orm.Get(session, exist, "runtime_name = ? and app_id = ? and service_name = ? and project_id = ? and workspace = ? and cluster_name = ?", dao.RuntimeName, dao.AppId, dao.ServiceName, dao.ProjectId, dao.Workspace, dao.ClusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if succ {
		return exist, nil
	}
	// check exist with table lock
	succ, err = orm.GetForUpdate(session, impl.engine, exist, "runtime_name = ? and app_id = ? and service_name = ? and project_id = ? and workspace = ? and cluster_name = ?", dao.RuntimeName, dao.AppId, dao.ServiceName, dao.ProjectId, dao.Workspace, dao.ClusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if succ {
		return exist, nil
	}
	_, err = orm.Insert(session, dao)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return dao, nil
}
