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

package testcase

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dbclient"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) BatchListAPIs(projectID uint64, tcIDs []uint64) (map[uint64][]*apistructs.ApiTestInfo, error) {
	begin := time.Now()
	m, err := dbclient.ListAPIsByTestCaseIDs(projectID, tcIDs)
	if err != nil {
		return nil, err
	}
	end := time.Now()
	logrus.Debugf("batch list APIs cost: %fs", end.Sub(begin).Seconds())

	r := make(map[uint64][]*apistructs.ApiTestInfo)
	for tcID := range m {
		for _, api := range m[tcID] {
			r[tcID] = append(r[tcID], convert2ReqStruct(api))
		}
	}
	return r, nil
}

func (svc *Service) ListAPIs(testCaseID int64) ([]*apistructs.ApiTestInfo, error) {
	apis, err := dbclient.GetApiTestListByUsecaseID(testCaseID)
	if err != nil {
		return nil, apierrors.ErrListAPITests.InternalError(err)
	}
	var results []*apistructs.ApiTestInfo
	for _, api := range apis {
		results = append(results, convert2ReqStruct(&api))
	}
	return results, nil
}

func (svc *Service) GetAPI(apiID int64) (*apistructs.ApiTestInfo, error) {
	api, err := dbclient.GetApiTest(apiID)
	if err != nil {
		return nil, apierrors.ErrGetAPITest.InternalError(err)
	}
	return convert2ReqStruct(api), nil
}

func (svc *Service) CreateAPI(req apistructs.ApiTestsCreateRequest) (*apistructs.ApiTestInfo, error) {
	req.Status = apistructs.ApiTestCreated
	api := convert2DBStruct(&req.ApiTestInfo)
	_, err := dbclient.CreateApiTest(api)
	if err != nil {
		return nil, apierrors.ErrCreateAPITest.InternalError(err)
	}
	return convert2ReqStruct(api), nil
}

func (svc *Service) UpdateAPI(req apistructs.ApiTestsUpdateRequest) (int64, error) {
	api := convert2DBStruct(&req.ApiTestInfo)

	if req.IsResult {
		if _, err := dbclient.UpdateApiTestResults(api); err != nil {
			return 0, apierrors.ErrUpdateAPITest.InternalError(err)
		}
	} else {
		if _, err := dbclient.UpdateApiTest(api); err != nil {
			return 0, apierrors.ErrUpdateAPITest.InternalError(err)
		}
	}

	return req.ApiID, nil
}

func (svc *Service) DeleteAPI(apiID int64) error {
	if err := dbclient.DeleteApiTest(apiID); err != nil {
		return apierrors.ErrDeleteAPITest.InternalError(err)
	}
	return nil
}

func convert2DBStruct(req *apistructs.ApiTestInfo) *dbclient.ApiTest {
	return &dbclient.ApiTest{
		ID:           req.ApiID,
		UsecaseID:    req.UsecaseID,
		UsecaseOrder: req.UsecaseOrder,
		ProjectID:    req.ProjectID,
		Status:       string(req.Status),
		ApiInfo:      req.ApiInfo,
		ApiRequest:   req.ApiRequest,
		ApiResponse:  req.ApiResponse,
		AssertResult: req.AssertResult,
	}
}
