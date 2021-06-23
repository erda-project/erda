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

// CreateNotifyHistory 创建通知发送历史
func (e *Endpoints) CreateNotifyHistory(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	if r.Body == nil {
		return apierrors.ErrCreateNotifyHistory.MissingParameter("body is nil").ToResp(), nil
	}

	var notifyHistoryCreateReq apistructs.CreateNotifyHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&notifyHistoryCreateReq); err != nil {
		return apierrors.ErrCreateNotifyHistory.InvalidParameter("can't decode body").ToResp(), nil
	}
	historyID, err := e.notifyGroup.CreateNotifyHistory(&notifyHistoryCreateReq)

	if err != nil {
		return apierrors.ErrCreateNotifyHistory.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(historyID)
}

// QueryNotifyHistories 查询通知发送历史
func (e *Endpoints) QueryNotifyHistories(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrQueryNotifyHistory.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	pageNo := getInt(r.URL, "pageNo", 1)
	pageSize := getInt(r.URL, "pageSize", 10)
	queryReq := apistructs.QueryNotifyHistoryRequest{
		PageSize:    pageSize,
		PageNo:      pageNo,
		Channel:     r.URL.Query().Get("channel"),
		NotifyName:  r.URL.Query().Get("notifyName"),
		StartTime:   r.URL.Query().Get("startTime"),
		EndTime:     r.URL.Query().Get("endTime"),
		Label:       r.URL.Query().Get("label"),
		ClusterName: r.URL.Query().Get("clusterName"),
		OrgID:       orgID,
	}
	result, err := e.notifyGroup.QueryNotifyHistories(&queryReq)
	if err != nil {
		return apierrors.ErrQueryNotifyHistory.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}
