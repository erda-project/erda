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

package code_coverage

import (
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func TestEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := monkey.Patch(GetJacocoAddr, func(uint64) string {
		return "addr"
	})
	defer m.Unpatch()

	bdl := NewMockCodeCoverageBDLer(ctrl)
	db := NewMockCodeCoverageDBer(ctrl)

	bdl.EXPECT().CheckPermission(gomock.Any()).Return(&apistructs.PermissionCheckResponseData{
		Access: true,
	}, nil)
	bdl.EXPECT().JacocoEnd(gomock.Any(), gomock.Any()).Return(nil)

	db.EXPECT().UpdateCodeCoverage(gomock.Any()).Return(nil)
	db.EXPECT().GetCodeCoverageByID(gomock.Any()).Return(&dao.CodeCoverageExecRecord{
		BaseModel: dbengine.BaseModel{},
		ProjectID: 1,
		Status:    apistructs.ReadyStatus,
	}, nil)

	svc := New(WithDBClient(db), WithBundle(bdl))
	if err := svc.End(apistructs.CodeCoverageUpdateRequest{
		ID: 0,
	}); err != nil {
		t.Error(err)
	}
}

func TestCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bdl := NewMockCodeCoverageBDLer(ctrl)
	db := NewMockCodeCoverageDBer(ctrl)

	bdl.EXPECT().CheckPermission(gomock.Any()).Return(&apistructs.PermissionCheckResponseData{
		Access: true,
	}, nil)

	db.EXPECT().CancelCodeCoverage(gomock.Any(), gomock.Any()).Return(nil)

	svc := New(WithDBClient(db), WithBundle(bdl))
	if err := svc.Cancel(apistructs.CodeCoverageCancelRequest{
		IdentityInfo: apistructs.IdentityInfo{},
		ProjectID:    1,
	}); err != nil {
		t.Error(err)
	}
}

func TestReadyCallBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().GetCodeCoverageByID(gomock.Any()).Return(&dao.CodeCoverageExecRecord{
		BaseModel: dbengine.BaseModel{},
		ProjectID: 1,
		Status:    apistructs.RunningStatus,
	}, nil)

	db.EXPECT().UpdateCodeCoverage(gomock.Any()).Return(nil)

	svc := New(WithDBClient(db))
	if err := svc.ReadyCallBack(apistructs.CodeCoverageUpdateRequest{
		IdentityInfo: apistructs.IdentityInfo{},
		ID:           1,
	}); err != nil {
		t.Error(err)
	}
}

func TestEndCallBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bdl := NewMockCodeCoverageBDLer(ctrl)
	db := NewMockCodeCoverageDBer(ctrl)

	bdl.EXPECT().GetProject(gomock.Any()).Return(&apistructs.ProjectDTO{
		ID:   0,
		Name: "foo",
	}, nil)

	db.EXPECT().GetCodeCoverageByID(gomock.Any()).Return(&dao.CodeCoverageExecRecord{
		BaseModel: dbengine.BaseModel{},
		ProjectID: 1,
		Status:    apistructs.EndingStatus,
	}, nil)
	db.EXPECT().UpdateCodeCoverage(gomock.Any()).Return(nil)

	svc := New(WithDBClient(db), WithBundle(bdl))
	if err := svc.EndCallBack(apistructs.CodeCoverageUpdateRequest{
		IdentityInfo:  apistructs.IdentityInfo{},
		ID:            1,
		Status:        "success",
		ReportXmlUUID: "",
	}); err != nil {
		t.Error(err)
	}
}

func TestReportCallBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().GetCodeCoverageByID(gomock.Any()).Return(&dao.CodeCoverageExecRecord{
		BaseModel:    dbengine.BaseModel{},
		ProjectID:    1,
		ReportStatus: apistructs.RunningStatus,
	}, nil)
	db.EXPECT().UpdateCodeCoverage(gomock.Any()).Return(nil)

	svc := New(WithDBClient(db))
	if err := svc.ReportCallBack(apistructs.CodeCoverageUpdateRequest{
		IdentityInfo: apistructs.IdentityInfo{},
		ID:           1,
		Status:       "success",
		ReportTarUrl: "bar",
	}); err != nil {
		t.Error(err)
	}
}

func TestListCodeCoverageRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().ListCodeCoverage(gomock.Any()).Return(nil, uint64(0), nil)

	svc := New(WithDBClient(db))
	if _, err := svc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{}); err != nil {
		t.Error(err)
	}
}

func TestGetCodeCoverageRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().GetCodeCoverageByID(gomock.Any()).Return(&dao.CodeCoverageExecRecord{
		BaseModel: dbengine.BaseModel{},
		ProjectID: 1,
		Status:    apistructs.ReadyStatus,
	}, nil)

	svc := New(WithDBClient(db))
	if _, err := svc.GetCodeCoverageRecord(1); err != nil {
		t.Error(err)
	}
}

func TestJudgeRunningRecordExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().ListCodeCoverageByStatus(gomock.Any(), gomock.Any()).Return(nil, nil)

	svc := New(WithDBClient(db))
	if err := svc.JudgeRunningRecordExist(1); err != nil {
		t.Error(err)
	}
}

func TestJudgeCanEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockCodeCoverageDBer(ctrl)

	db.EXPECT().ListCodeCoverageByStatus(gomock.Any(), gomock.Any()).Return(nil, nil)

	svc := New(WithDBClient(db))
	if _, err := svc.JudgeCanEnd(1); err != nil {
		t.Error(err)
	}
}
