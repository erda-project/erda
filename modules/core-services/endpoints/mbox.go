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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// QueryMBox 查询站内信
func (e *Endpoints) QueryMBox(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrQueryMBox.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	pageNo := getInt(r.URL, "pageNo", 1)
	pageSize := getInt(r.URL, "pageSize", 10)
	queryReq := &apistructs.QueryMBoxRequest{
		PageSize: pageSize,
		PageNo:   pageNo,
		Label:    r.URL.Query().Get("label"),
		UserID:   r.Header.Get("User-ID"),
		OrgID:    orgID,
	}
	result, err := e.mbox.QueryMBox(queryReq)
	if err != nil {
		return apierrors.ErrQueryMBox.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// GetMBox 获取站内信详情
func (e *Endpoints) GetMBox(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrQueryMBox.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	idStr := vars["mboxID"]
	mboxID, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		return apierrors.ErrQueryMBox.InvalidParameter(err).ToResp(), nil
	}

	result, err := e.mbox.GetMBox(mboxID, orgID, r.Header.Get("User-ID"))
	if err != nil {
		return apierrors.ErrQueryMBox.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// CreateMBox 创建站内信 内部接口
func (e *Endpoints) CreateMBox(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	createReq := &apistructs.CreateMBoxRequest{}
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateMBox.InvalidParameter("can't decode body").ToResp(), nil
	}
	err := e.mbox.CreateMBox(createReq)
	if err != nil {
		return apierrors.ErrCreateMBox.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("")
}

// GetMBoxStats 查询站内信统计信息
func (e *Endpoints) GetMBoxStats(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetMBoxStats.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	mboxStats, err := e.mbox.GetMBoxStats(orgID, r.Header.Get("User-ID"))
	if err != nil {
		return apierrors.ErrGetMBoxStats.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(mboxStats)
}

// SetMBoxReadStatus 设置站内信已经读标记
func (e *Endpoints) SetMBoxReadStatus(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrSetMBoxReadStatus.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	req := &apistructs.SetMBoxReadStatusRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrSetMBoxReadStatus.InvalidParameter("can't decode body").ToResp(), nil
	}
	req.OrgID = orgID
	req.UserID = r.Header.Get("User-ID")
	err = e.mbox.SetMBoxReadStatus(req)
	if err != nil {
		return apierrors.ErrSetMBoxReadStatus.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}
