// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"github.com/pkg/errors"
	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type GatewayHubInfoServiceImpl struct {
	*SessionHelper

	engine   *orm.OrmEngine
	executor xorm.Interface
}

func NewGatewayHubInfoServiceImpl() (GatewayHubInfoService, error) {
	engine, err := orm.GetSingleton()
	if err != nil {
		return nil, errors.Wrap(err, "new GatewayUpstreamServiceImpl failed")
	}
	return &GatewayHubInfoServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayHubInfoServiceImpl) NewSession(helper ...*SessionHelper) (GatewayHubInfoService, error) {
	var (
		session *SessionHelper
		err     error
	)
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayHubInfoServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil
	} else {
		session = helper[0]
	}
	return &GatewayHubInfoServiceImpl{
		SessionHelper: session,
		engine:        impl.engine,
		executor:      session.session,
	}, nil
}

func (impl *GatewayHubInfoServiceImpl) GetByAny(cond *orm.GatewayHubInfo) (*orm.GatewayHubInfo, bool, error) {
	if cond == nil {
		return nil, false, errors.New(vars.ERR_INVALID_ARG)
	}
	var result orm.GatewayHubInfo
	conds, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to BuildConds")
	}
	success, err := orm.GetByAnyI(impl.executor, conds, &result)
	if err != nil {
		return nil, false, errors.Wrap(err, vars.ERR_SQL_FAIL)
	}
	if !success {
		return nil, false, nil
	}
	return &result, true, nil
}

func (impl *GatewayHubInfoServiceImpl) SelectByAny(cond *orm.GatewayHubInfo) ([]orm.GatewayHubInfo, error) {
	var result []orm.GatewayHubInfo
	if cond == nil {
		return nil, errors.New(vars.ERR_INVALID_ARG)
	}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, vars.ERR_SQL_FAIL)
	}
	err = orm.SelectByAnyI(impl.executor.Asc("create_time"), bCond, &result)
	if err != nil {
		return nil, errors.Wrap(err, vars.ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayHubInfoServiceImpl) Insert(dao *orm.GatewayHubInfo) error {
	if dao == nil {
		return errors.New(vars.ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.executor, dao)
	if err != nil {
		return errors.Wrap(err, vars.ERR_SQL_FAIL)
	}
	return nil
}
