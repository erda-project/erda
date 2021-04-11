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

package testset

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) Copy(req apistructs.TestSetCopyRequest) (uint64, error) {
	// 参数校验
	if req.TestSetID == 0 {
		return 0, apierrors.ErrCopyTestSet.InvalidParameter("cannot copy root testset")
	}
	if req.CopyToTestSetID == req.TestSetID {
		return 0, apierrors.ErrCopyTestSet.InvalidParameter("cannot copy to itself")
	}

	// 查询待拷贝测试集
	srcTs, err := svc.Get(req.TestSetID)
	if err != nil {
		return 0, err
	}

	// 查询目标测试集
	var dstTs *apistructs.TestSet
	if req.CopyToTestSetID != 0 {
		dstTs, err = svc.Get(req.CopyToTestSetID)
		if err != nil {
			return 0, err
		}
		findInSub, err := svc.findTargetTestSetIDInSubTestSets([]uint64{srcTs.ID}, srcTs.ProjectID, dstTs.ID)
		if err != nil {
			return 0, apierrors.ErrCopyTestSet.InternalError(err)
		}
		if findInSub {
			return 0, apierrors.ErrCopyTestSet.InvalidParameter("cannot copy to sub testset")
		}
	}

	// 遍历测试集做拷贝
	copiedTsIDs, err := svc.recursiveCopy(srcTs, dstTs, req.IdentityInfo)
	if err != nil {
		return 0, err
	}
	return copiedTsIDs[0], nil
}

func (svc *Service) recursiveCopy(srcTs, dstTs *apistructs.TestSet, identityInfo apistructs.IdentityInfo) ([]uint64, error) {
	var newTsIDs []uint64
	if dstTs == nil {
		dstTs = &apistructs.TestSet{
			ID:        0,
			ProjectID: srcTs.ProjectID,
			ParentID:  0,
			Recycled:  srcTs.Recycled,
			Directory: "/",
			Order:     0,
		}
	}
	// 先创建新测试集
	copiedTs, err := svc.Create(apistructs.TestSetCreateRequest{
		ProjectID: &srcTs.ProjectID,
		ParentID:  &dstTs.ID,
		Name:      srcTs.Name,
	})
	if err != nil {
		return nil, err
	}
	newTsIDs = append(newTsIDs, copiedTs.ID)

	// 将 原测试集下的用例复制到新测试集
	_, waitCopyTcIDs, err := svc.tcSvc.ListTestCases(apistructs.TestCaseListRequest{
		ProjectID:  srcTs.ProjectID,
		TestSetIDs: []uint64{srcTs.ID},
		Recycled:   false,
		IDOnly:     true,
	})
	if err != nil {
		return nil, err
	}
	if len(waitCopyTcIDs) > 0 {
		_, err = svc.tcSvc.BatchCopyTestCases(apistructs.TestCaseBatchCopyRequest{
			CopyToTestSetID: copiedTs.ID,
			ProjectID:       copiedTs.ProjectID,
			TestCaseIDs:     waitCopyTcIDs,
			IdentityInfo:    identityInfo,
		})
		if err != nil {
			return nil, err
		}
	}

	// 递归调用子测试集
	subTestSets, err := svc.List(apistructs.TestSetListRequest{
		Recycled:  false,
		ParentID:  &srcTs.ID,
		ProjectID: &srcTs.ProjectID,
	})
	if err != nil {
		return nil, err
	}
	for _, subTs := range subTestSets {
		newSubTsIDs, err := svc.recursiveCopy(&subTs, copiedTs, identityInfo)
		if err != nil {
			return nil, err
		}
		newTsIDs = append(newTsIDs, newSubTsIDs...)
	}
	return newTsIDs, nil
}
