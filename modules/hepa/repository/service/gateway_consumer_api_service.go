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

type GatewayConsumerApiServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayConsumerApiServiceImpl() (*GatewayConsumerApiServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayConsumerApiServiceImpl failed")
	}
	return &GatewayConsumerApiServiceImpl{engine}, nil
}

// func (impl *GatewayConsumerApiServiceImpl) InsertOrUpdate(api *orm.GatewayConsumerApi) error {
// 	if api == nil {
// 		return errors.New(ERR_INVALID_ARG)
// 	}
// 	engine, err := impl.engine.GetEngine()
// 	if err != nil {
// 		return errors.Wrap(err, "GetEngine failed")
// 	}
// 	session := engine.NewSession()
// 	defer session.Close()
// 	err = session.Begin()
// 	exist := &orm.GatewayConsumerApi{}
// 	succ, err := session.Get(exist)
// 	if err != nil {
// 		return errors.Wrap(err, ERR_SQL_FAIL)
// 	}
// 	if succ {
// 		affected, err := session.Id(exist.Id).Update(api)
// 		if err != nil {
// 			return errors.Wrap(err, ERR_SQL_FAIL)
// 		}
// 		if affected < 1 {
// 			return errors.New(ERR_NO_CHANGE)
// 		}
// 		api.Id = exist.Id
// 		return session.Commit()
// 	}
// 	affected, err := session.Insert(api)
// 	if err != nil {
// 		return errors.Wrap(err, ERR_SQL_FAIL)
// 	}
// 	if affected < 1 {
// 		return errors.New(ERR_NO_CHANGE)
// 	}
// 	return session.Commit()
// }

func (impl *GatewayConsumerApiServiceImpl) Update(api *orm.GatewayConsumerApi) error {
	if api == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, api)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayConsumerApiServiceImpl) Insert(api *orm.GatewayConsumerApi) error {
	if api == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(impl.engine, api)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayConsumerApiServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayConsumerApi{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	// if affected < 1 {
	// 	return errors.New(ERR_NO_CHANGE)
	// }
	return nil

}

func (impl *GatewayConsumerApiServiceImpl) GetById(id string) (*orm.GatewayConsumerApi, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	ConsumerApi := &orm.GatewayConsumerApi{}
	succ, err := orm.Get(impl.engine, ConsumerApi, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return ConsumerApi, nil
}

func (impl *GatewayConsumerApiServiceImpl) GetByConsumerAndApi(consumerId string, apiId string) (*orm.GatewayConsumerApi, error) {
	if len(apiId) == 0 || len(consumerId) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	api := &orm.GatewayConsumerApi{}
	succ, err := orm.Get(impl.engine, api, "consumer_id = ? and api_id = ?",
		consumerId, apiId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return api, nil
}

func (impl *GatewayConsumerApiServiceImpl) SelectByConsumer(consumerId string) ([]orm.GatewayConsumerApi, error) {
	var result []orm.GatewayConsumerApi
	if len(consumerId) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.engine.Desc("create_time"), &result, "consumer_id = ?", consumerId)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}

func (impl *GatewayConsumerApiServiceImpl) SelectByApi(apiId string) ([]orm.GatewayConsumerApi, error) {
	var result []orm.GatewayConsumerApi
	if len(apiId) == 0 {
		return result, errors.New(ERR_INVALID_ARG)
	}
	err := orm.Select(impl.engine.Desc("create_time"), &result, "api_id = ?", apiId)
	if err != nil {
		return result, errors.Wrap(err, ERR_SQL_FAIL)
	}
	return result, nil
}
