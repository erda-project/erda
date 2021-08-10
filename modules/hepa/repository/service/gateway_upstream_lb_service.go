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

type GatewayUpstreamLbServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayUpstreamLbServiceImpl() (*GatewayUpstreamLbServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayUpstreamLbServiceImpl failed")
	}
	return &GatewayUpstreamLbServiceImpl{engine}, nil
}

func (impl GatewayUpstreamLbServiceImpl) Get(cond *orm.GatewayUpstreamLb) (*orm.GatewayUpstreamLb, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayUpstreamLb{}
	succ, err := orm.GetByAny(impl.engine, dao, cond)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl GatewayUpstreamLbServiceImpl) GetForUpdate(session *xorm.Session, cond *orm.GatewayUpstreamLb) (*orm.GatewayUpstreamLb, error) {
	if cond == nil || cond.OrgId == "" || cond.ProjectId == "" || cond.Env == "" || cond.Az == "" || cond.LbName == "" {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayUpstreamLb{}
	succ, err := orm.GetForUpdate(session, impl.engine, dao, "org_id = ? and project_id = ? and env = ? and az = ? and lb_name = ?", cond.OrgId, cond.ProjectId, cond.Env, cond.Az, cond.LbName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl GatewayUpstreamLbServiceImpl) Insert(session *xorm.Session, dao *orm.GatewayUpstreamLb) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(session, dao)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (impl GatewayUpstreamLbServiceImpl) UpdateDeploymentId(id string, deploymentId int) error {
	if id == "" {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, &orm.GatewayUpstreamLb{
		BaseRow:          orm.BaseRow{Id: id},
		LastDeploymentId: deploymentId,
	}, "last_deployment_id")
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl GatewayUpstreamLbServiceImpl) GetById(id string) (*orm.GatewayUpstreamLb, error) {
	if id == "" {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayUpstreamLb{}
	succ, err := orm.Get(impl.engine, dao, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl GatewayUpstreamLbServiceImpl) GetByKongId(id string) (*orm.GatewayUpstreamLb, error) {
	if id == "" {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayUpstreamLb{}
	succ, err := orm.Get(impl.engine, dao, "kong_upstream_id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}
