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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// CreateNotifyGroup 创建通知组
func (e *Endpoints) CreateNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	locale := e.GetLocale(r)
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrCreateNotifyGroup.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	if r.Body == nil {
		return apierrors.ErrCreateNotifyGroup.MissingParameter("body is nil").ToResp(), nil
	}

	var notifyGroupCreateReq apistructs.CreateNotifyGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&notifyGroupCreateReq); err != nil {
		return apierrors.ErrCreateNotifyGroup.InvalidParameter("can't decode body").ToResp(), nil
	}
	if strings.TrimSpace(notifyGroupCreateReq.Name) == "" {
		return apierrors.ErrCreateNotifyGroup.InvalidParameter("name is empty").ToResp(), nil
	}
	if utf8.RuneCountInString(notifyGroupCreateReq.Name) > 50 {
		return apierrors.ErrCreateNotifyGroup.InvalidParameter(locale.Get("ErrCreateNotifyGroup.NameTooLong")).ToResp(), nil
	}

	err = e.checkNotifyPermission(r, notifyGroupCreateReq.ScopeType, notifyGroupCreateReq.ScopeID, apistructs.CreateAction)
	if err != nil {
		return apierrors.ErrCreateNotifyGroup.AccessDenied().ToResp(), nil
	}

	notifyGroupCreateReq.Creator = r.Header.Get("User-Id")
	notifyGroupCreateReq.OrgID = orgID
	notifyGroupID, err := e.notifyGroup.Create(locale, &notifyGroupCreateReq)

	if err != nil {
		return apierrors.ErrCreateNotifyGroup.InternalError(err).ToResp(), nil
	}

	notifyGroup, err := e.notifyGroup.Get(notifyGroupID, orgID)
	if err != nil {
		return apierrors.ErrCreateNotifyGroup.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(notifyGroup)
}

// UpdateNotifyGroup 更新通知组
func (e *Endpoints) UpdateNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	idStr := vars["groupID"]
	notifyGroupId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdateNotifyGroup.InvalidParameter(err).ToResp(), nil
	}

	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrUpdateNotify.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrUpdateNotifyGroup.MissingParameter("body is nil").ToResp(), nil
	}

	var notifyGroupUpdateReq apistructs.UpdateNotifyGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&notifyGroupUpdateReq); err != nil {
		return apierrors.ErrUpdateNotifyGroup.InvalidParameter("can't decode body").ToResp(), nil
	}
	notifyGroupUpdateReq.ID = notifyGroupId
	notifyGroupUpdateReq.OrgID = orgID

	gronotifyGroup, err := e.notifyGroup.Get(notifyGroupId, orgID)
	if err != nil {
		return apierrors.ErrUpdateNotifyGroup.InternalError(err).ToResp(), nil
	}
	err = e.checkNotifyPermission(r, gronotifyGroup.ScopeType, gronotifyGroup.ScopeID, apistructs.UpdateAction)
	if err != nil {
		return apierrors.ErrUpdateNotifyGroup.AccessDenied().ToResp(), nil
	}

	err = e.notifyGroup.Update(&notifyGroupUpdateReq)
	if err != nil {
		return apierrors.ErrUpdateNotifyGroup.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(gronotifyGroup)
}

// GetNotifyGroup 获取通知组信息
func (e *Endpoints) GetNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	idStr := vars["groupID"]
	notifyGroupId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.InvalidParameter(err).ToResp(), nil
	}
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	notifyGroup, err := e.notifyGroup.Get(notifyGroupId, orgID)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.InternalError(err).ToResp(), nil
	}
	err = e.checkNotifyPermission(r, notifyGroup.ScopeType, notifyGroup.ScopeID, apistructs.GetAction)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.AccessDenied().ToResp(), nil
	}
	var userIDs []string
	if notifyGroup.Creator != "" {
		userIDs = append(userIDs, notifyGroup.Creator)
	}
	for _, target := range notifyGroup.Targets {
		if target.Type == apistructs.UserNotifyTarget {
			for _, t := range target.Values {
				userIDs = append(userIDs, t.Receiver)
			}
		}
	}
	return httpserver.OkResp(notifyGroup, userIDs)
}

// GetNotifyGroupDetail 获取通知组信息详情
func (e *Endpoints) GetNotifyGroupDetail(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	idStr := vars["groupID"]
	notifyGroupId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.InvalidParameter(err).ToResp(), nil
	}

	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	notifyGroup, err := e.notifyGroup.GetDetail(notifyGroupId, orgID)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.InternalError(err).ToResp(), nil
	}
	err = e.checkNotifyPermission(r, notifyGroup.ScopeType, notifyGroup.ScopeID, apistructs.GetAction)
	if err != nil {
		return apierrors.ErrGetNotifyGroup.AccessDenied().ToResp(), nil
	}

	return httpserver.OkResp(notifyGroup)
}

// QueryNotifyGroup 查询通知组
func (e *Endpoints) QueryNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrQueryNotifyGroup.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	pageNo := getInt(r.URL, "pageNo", 1)
	pageSize := getInt(r.URL, "pageSize", 10)
	queryReq := apistructs.QueryNotifyGroupRequest{
		PageSize:  pageSize,
		PageNo:    pageNo,
		ScopeType: r.URL.Query().Get("scopeType"),
		ScopeID:   r.URL.Query().Get("scopeId"),
		Label:     r.URL.Query().Get("label"),
	}
	err = e.checkNotifyPermission(r, queryReq.ScopeType, queryReq.ScopeID, apistructs.ListAction)
	if err != nil {
		return apierrors.ErrQueryNotifyGroup.AccessDenied().ToResp(), nil
	}
	result, err := e.notifyGroup.Query(&queryReq, orgID)
	if err != nil {
		return apierrors.ErrQueryNotifyGroup.InternalError(err).ToResp(), nil
	}
	var userIDs []string
	for _, group := range result.List {
		userIDs = append(userIDs, group.Creator)
		for _, target := range group.Targets {
			if target.Type == apistructs.UserNotifyTarget {
				for _, t := range target.Values {
					userIDs = append(userIDs, t.Receiver)
				}
			}
		}
	}
	return httpserver.OkResp(result, userIDs)
}

// BatchGetNotifyGroup 批量根据id查询通知组 内部接口
func (e *Endpoints) BatchGetNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	idsStr := r.URL.Query().Get("ids")
	var ids []int64
	for _, idStr := range strings.Split(idsStr, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	result, err := e.notifyGroup.BatchGet(ids)
	if err != nil {
		return apierrors.ErrQueryNotifyGroup.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(result)
}

// DeleteNotifyGroup 删除通知组
func (e *Endpoints) DeleteNotifyGroup(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotifyGroup.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	idStr := vars["groupID"]
	notifyGroupId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotifyGroup.InvalidParameter(err).ToResp(), nil
	}

	notifyGroup, err := e.notifyGroup.Get(notifyGroupId, orgID)
	if err != nil {
		return apierrors.ErrDeleteNotifyGroup.InternalError(err).ToResp(), nil
	}
	err = e.checkNotifyPermission(r, notifyGroup.ScopeType, notifyGroup.ScopeID, apistructs.DeleteAction)
	if err != nil {
		return apierrors.ErrDeleteNotifyGroup.AccessDenied().ToResp(), nil
	}

	if err = e.notifyGroup.Delete(notifyGroupId, orgID); err != nil {
		return apierrors.ErrDeleteNotifyGroup.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(notifyGroup)
}

func getInt(url *url.URL, key string, defaultValue int64) int64 {
	valueStr := url.Query().Get(key)
	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return defaultValue
	}
	return value
}
