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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateTestSet 创建测试集
func (e *Endpoints) CreateTestSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateTestSet.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestSet.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestSetCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateTestSet.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// create
	result, err := e.testset.Create(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) GetTestSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetTestSet.NotLogin().ToResp(), nil
	}
	_ = identityInfo

	testSetID, err := strconv.ParseUint(vars["testSetID"], 10, 64)
	if err != nil {
		logrus.Errorf("failed to parse testSetID, input: %s, err: %v", vars["testSetID"], err)
		return apierrors.ErrGetTestSet.InvalidParameter("testSetID").ToResp(), nil
	}

	ts, err := e.testset.Get(testSetID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(ts, strutil.DedupSlice([]string{ts.CreatorID, ts.UpdaterID}))
}

// ListTestSets 获取测试集列表
func (e *Endpoints) ListTestSets(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.TestSetListRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListTestSets.InvalidParameter(err).ToResp(), nil
	}

	testSets, err := e.testset.List(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testSets)
}

// UpdateTestSet 更新测试集
func (e *Endpoints) UpdateTestSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestSet.NotLogin().ToResp(), nil
	}

	var req apistructs.TestSetUpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestSet.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestSet.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if err := e.testset.Update(req); err != nil {
		return errorresp.ErrResp(err)
	}

	ts, err := e.testset.Get(req.TestSetID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(ts)
}

// RecycleTestSet 回收测试集
func (e *Endpoints) RecycleTestSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrRecycleTestSet.NotLogin().ToResp(), nil
	}

	testSetID, err := strconv.ParseUint(vars["testSetID"], 10, 64)
	if err != nil {
		logrus.Errorf("failed to parse testSetID from path, value: %s, err: %v", vars["testSetID"], err)
		return apierrors.ErrRecycleTestSet.InvalidParameter("testSetID").ToResp(), nil
	}

	if err := e.testset.Recycle(apistructs.TestSetRecycleRequest{
		TestSetID:    testSetID,
		IsRoot:       true,
		IdentityInfo: identityInfo,
	}); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) CleanTestSetFromRecycleBin(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCleanTestSetFromRecycleBin.NotLogin().ToResp(), nil
	}

	testSetID, err := strconv.ParseUint(vars["testSetID"], 10, 64)
	if err != nil {
		logrus.Errorf("failed to parse testSetID from path, value: %s, err: %v", vars["testSetID"], err)
		return apierrors.ErrCleanTestSetFromRecycleBin.InvalidParameter("testSetID").ToResp(), nil
	}

	if err := e.testset.CleanFromRecycleBin(apistructs.TestSetCleanFromRecycleBinRequest{
		TestSetID:    testSetID,
		IdentityInfo: identityInfo,
	}); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// CleanTestSet 清空测试集
func (e *Endpoints) CleanTestSet(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateTestSet.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 64)
	if err != nil {
		return apierrors.ErrUpdateTestSet.InvalidParameter(err).ToResp(), nil
	}

	projectIDStr := r.URL.Query().Get("projectId")
	if projectIDStr == "" {
		return apierrors.ErrListTestSets.MissingParameter("projectId").ToResp(), nil
	}
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 64)

	testsetIDStr := vars["id"]
	if testsetIDStr == "" {
		return apierrors.ErrListTestSets.MissingParameter("id").ToResp(), nil
	}
	testsetID, _ := strconv.ParseUint(testsetIDStr, 10, 64)

	if err := e.testset.Clean(userID, testsetID, projectID); err != nil {
		return apierrors.ErrUpdateTestSet.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("update success")
}

// RecoverTestSetFromRecycleBin 从回收站恢复测试集（递归子测试集和测试用例）
func (e *Endpoints) RecoverTestSetFromRecycleBin(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrRecoverTestSetFromRecycleBin.NotLogin().ToResp(), nil
	}

	testSetID, err := strconv.ParseUint(vars["testSetID"], 10, 64)
	if err != nil {
		logrus.Errorf("failed to parse testSetID, input: %s, err: %v", vars["testSetID"], err)
		return apierrors.ErrRecoverTestSetFromRecycleBin.InvalidParameter("testSetID").ToResp(), nil
	}

	var req apistructs.TestSetRecoverFromRecycleBinRequest
	if r.ContentLength == 0 {
		return apierrors.ErrRecoverTestSetFromRecycleBin.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrRecoverTestSetFromRecycleBin.InvalidParameter(err).ToResp(), nil
	}
	req.TestSetID = testSetID
	req.IdentityInfo = identityInfo

	if err := e.testset.RecoverFromRecycleBin(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// RestoreTestSet 回收站恢复测试集
func (e *Endpoints) RecoverTestSet(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateTestSet.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 64)
	if err != nil {
		return apierrors.ErrUpdateTestSet.InvalidParameter(err).ToResp(), nil
	}

	projectIDStr := r.URL.Query().Get("projectId")
	if projectIDStr == "" {
		return apierrors.ErrListTestSets.MissingParameter("projectId").ToResp(), nil
	}
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 64)

	testsetIDStr := vars["id"]
	if testsetIDStr == "" {
		return apierrors.ErrListTestSets.MissingParameter("id").ToResp(), nil
	}
	testsetID, _ := strconv.ParseUint(testsetIDStr, 10, 64)

	targetTestSetIDStr := r.URL.Query().Get("targetTestSetId")
	if testsetIDStr == "" {
		return apierrors.ErrListTestSets.MissingParameter("targetTestSetId").ToResp(), nil
	}
	targetTestSetID, _ := strconv.ParseUint(targetTestSetIDStr, 10, 64)

	if err := e.testset.Recover(userID, testsetID, projectID, targetTestSetID); err != nil {
		return apierrors.ErrUpdateTestSet.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("update success")
}

// CopyTestSet 拷贝测试集
func (e *Endpoints) CopyTestSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCopyTestSet.NotLogin().ToResp(), nil
	}

	testSetID, err := strconv.ParseUint(vars["testSetID"], 10, 64)
	if err != nil {
		logrus.Errorf("failed to parse testSetID, input: %s, err: %v", vars["testSetID"], err)
		return apierrors.ErrCopyTestSet.InvalidParameter("testSetID").ToResp(), nil
	}

	var req apistructs.TestSetCopyRequest
	if r.ContentLength == 0 {
		return apierrors.ErrCopyTestSet.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCopyTestSet.InvalidParameter(err).ToResp(), nil
	}
	req.TestSetID = testSetID
	req.IdentityInfo = identityInfo

	copiedTestSetID, err := e.testset.Copy(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(copiedTestSetID)
}
