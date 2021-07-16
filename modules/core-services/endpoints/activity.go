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
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// ListActivity 活动列表
func (e *Endpoints) ListActivity(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取请求参数
	param, err := getActivityListParam(r)
	if err != nil {
		return apierrors.ErrListActivity.InvalidParameter(err).ToResp(), nil
	}
	if param.RuntimeID == 0 {
		return apierrors.ErrListActivity.MissingParameter("runtime id").ToResp(), nil
	}
	total, activities, err := e.activity.ListByRuntime(param.RuntimeID, param.PageNo, param.PageSize)
	if err != nil {
		return apierrors.ErrListActivity.InternalError(err).ToResp(), nil
	}

	activityDTOs := make([]apistructs.ActivityDTO, 0, len(activities))
	userIDs := make([]string, 0, len(activities))
	for i := range activities {
		activityDTOs = append(activityDTOs, *convertToActivityDTO(&activities[i]))
		userIDs = append(userIDs, activities[i].UserID)
	}

	return httpserver.OkResp(apistructs.ActivityListResponseData{Total: total, List: activityDTOs}, userIDs)
}

func getActivityListParam(r *http.Request) (*apistructs.ActivitiyListRequest, error) {
	// 获取runtimeID
	runtimeIDStr := r.URL.Query().Get("runtimeId")
	var (
		runtimeID int64
		err       error
	)
	if runtimeIDStr != "" {
		runtimeID, err = strutil.Atoi64(runtimeIDStr)
		if err != nil {
			return nil, err
		}
	}

	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, err
	}

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, err
	}

	return &apistructs.ActivitiyListRequest{
		RuntimeID: runtimeID,
		PageNo:    pageNo,
		PageSize:  pageSize,
	}, nil
}

func convertToActivityDTO(activity *model.Activity) *apistructs.ActivityDTO {
	return &apistructs.ActivityDTO{
		ID:            activity.ID,
		OrgID:         activity.OrgID,
		ProjectID:     activity.ProjectID,
		ApplicationID: activity.ApplicationID,
		RuntimeID:     activity.RuntimeID,
		UserID:        activity.UserID,
		Type:          activity.Type,
		Action:        activity.Action,
		Desc:          activity.Desc,
		Context:       activity.Context,
		CreatedAt:     activity.CreatedAt,
	}
}
