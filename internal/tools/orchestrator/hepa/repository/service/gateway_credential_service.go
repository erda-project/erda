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

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type GatewayCredentialServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayCredentialServiceImpl() (*GatewayCredentialServiceImpl, error) {
	engine, err := orm.GetSingleton()
	if err != nil {
		return nil, errors.Wrap(err, "new GatewayConsumerServiceImpl failed")
	}
	return &GatewayCredentialServiceImpl{engine}, nil
}

func (impl *GatewayCredentialServiceImpl) Insert(credential *orm.GatewayCredential) error {
	if credential == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, credential)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayCredentialServiceImpl) Update(credential *orm.GatewayCredential) error {
	if credential == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, credential)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayCredentialServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayCredential{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayCredentialServiceImpl) DeleteByConsumerId(consumerId string) error {
	if len(consumerId) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayCredential{}, "consumer_id = ?", consumerId)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayCredentialServiceImpl) GetByConsumerAndApi(apiId string, consumerId string) (*orm.GatewayCredential, error) {
	if len(apiId) == 0 || len(consumerId) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	credential := &orm.GatewayCredential{}
	succ, err := orm.Get(impl.engine, credential, "consumer_id = ? and api_id = ?",
		consumerId, apiId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return credential, nil
}

func (impl *GatewayCredentialServiceImpl) GetById(id string) (*orm.GatewayCredential, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	credential := &orm.GatewayCredential{}
	succ, err := orm.Get(impl.engine, credential, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return credential, nil
}

func (impl *GatewayCredentialServiceImpl) GetByConsumerId(consumerId string) (*orm.GatewayCredential, error) {
	if len(consumerId) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	credential := &orm.GatewayCredential{}
	succ, err := orm.Get(impl.engine, credential, "ConsumerId = ?", consumerId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return credential, nil
}

func (impl *GatewayCredentialServiceImpl) SelectByConsumerId(consumerId string) ([]orm.GatewayCredential, error) {
	var result []orm.GatewayCredential
	if len(consumerId) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.engine.Desc("create_time"), &result, "consumer_id = ?", consumerId)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}
