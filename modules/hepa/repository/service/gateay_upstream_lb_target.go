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
	log "github.com/sirupsen/logrus"

	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GatewayUpstreamLbTargetServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayUpstreamLbTargetServiceImpl() (*GatewayUpstreamLbTargetServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayUpstreamLbTargetServiceImpl failed")
	}
	return &GatewayUpstreamLbTargetServiceImpl{engine}, nil
}

func (impl GatewayUpstreamLbTargetServiceImpl) Insert(dao *orm.GatewayUpstreamLbTarget) error {
	if dao == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, dao)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (impl GatewayUpstreamLbTargetServiceImpl) Select(lbId, target string) ([]orm.GatewayUpstreamLbTarget, error) {
	if lbId == "" || target == "" {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	var result []orm.GatewayUpstreamLbTarget
	err := orm.Select(impl.engine.Desc("create_time"), &result, "lb_id = ? and target = ?", lbId, target)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (impl GatewayUpstreamLbTargetServiceImpl) Delete(id string) error {
	if id == "" {
		return errors.New(ERR_INVALID_ARG)
	}
	affected, err := orm.Delete(impl.engine, &orm.GatewayUpstreamLbTarget{}, "id = ?", id)
	if err != nil {
		return errors.WithStack(err)
	}
	if affected < 1 {
		log.Debugf("%+v", errors.New("maybe already deleted"))
	}
	return nil
}

func (impl GatewayUpstreamLbTargetServiceImpl) SelectByDeploymentId(id int) ([]orm.GatewayUpstreamLbTarget, error) {
	var result []orm.GatewayUpstreamLbTarget
	err := orm.Select(impl.engine, &result, "deployment_id = ?", id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}
