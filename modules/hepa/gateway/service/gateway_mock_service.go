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
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
)

type GatewayMockServiceImpl struct {
	mockDb db.GatewayMockService
}

func NewGatewayMockServiceImpl() (*GatewayMockServiceImpl, error) {
	mockDb, err := db.NewGatewayMockServiceImpl()
	if err != nil {
		return nil, errors.Wrap(err, "NewGatewayMockServiceImpl failed")
	}
	return &GatewayMockServiceImpl{
		mockDb: mockDb,
	}, nil
}

func (impl GatewayMockServiceImpl) RegisterMockApi(dto *gw.MockInfoDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	mock := &orm.GatewayMock{
		Az:      dto.Az,
		Body:    dto.Body,
		HeadKey: dto.HeadKey,
		Pathurl: dto.PathUrl,
		Method:  dto.Method,
	}
	exist, err := impl.mockDb.GetMockByAny(&orm.GatewayMock{
		HeadKey: dto.HeadKey,
	})
	if err != nil {
		log.Error(errors.WithStack(err))
		return res
	}
	if exist == nil {
		err = impl.mockDb.Insert(mock)
		if err != nil {
			log.Error(errors.WithStack(err))
			return res
		}
		return res.SetSuccessAndData(true)
	}
	mock.Id = exist.Id
	err = impl.mockDb.Update(mock)
	if err != nil {
		log.Error(errors.WithStack(err))
		return res
	}
	return res.SetSuccessAndData(true)
}

func (impl GatewayMockServiceImpl) CallMockApi(headKey, pathUrl, method string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	mock, err := impl.mockDb.GetMockByAny(&orm.GatewayMock{
		HeadKey: headKey,
		Pathurl: pathUrl,
		Method:  method,
	})
	if err != nil {
		log.Error(errors.WithStack(err))
		return res
	}
	if mock == nil {
		log.Error(errors.WithStack(err))
		return res.SetReturnCode(MOCK_IS_NOT_EXISTS)
	}
	data := mock.Body
	return res.SetSuccessAndData(data)
}
