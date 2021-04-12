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
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

// IssueCallback 事件管理 hook的回调
func (e *Endpoints) IssueCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req apistructs.IssueEvent
	)
	if r.Body == nil {
		return apierrors.ErrIssueCallback.MissingParameter("body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrIssueCallback.InvalidParameter(err).ToResp(), nil
	}

	marshal, _ := json.Marshal(req)

	logrus.Printf("action:%s StreamType:%s StreamParams:%s title:%s content:%s", req.Action, req.Content.StreamType, string(marshal), req.Content.Title, req.Content.Content)

	err := e.processIssueEvent(req)
	if err != nil {
		logrus.Errorf("failed to process issue event, (%+v)", err)
		return apierrors.ErrIssueCallback.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp("")
}

func (e *Endpoints) processIssueEvent(req apistructs.IssueEvent) error {
	eventnName := "issue_" + req.Action
	notifyDetails, err := e.bdl.QueryNotifiesBySource(req.OrgID, "project", req.ProjectID, eventnName, "")
	if err != nil {
		return err
	}
	orgID, err := strconv.ParseInt(req.OrgID, 10, 64)
	if err != nil {
		return err
	}
	for _, notifyDetail := range notifyDetails {
		if notifyDetail.NotifyGroup == nil {
			continue
		}
		notifyItem := notifyDetail.NotifyItems[0]
		params := map[string]string{
			"issue_title": req.Content.Title,
			"content":     req.Content.Content,
			"atUserIDs":   req.Content.AtUserIDs,
		}
		marshal, _ := json.Marshal(params)
		logrus.Infof("issue params :%s", string(marshal))

		// 不使用notifyItemID是为了支持第三方通知项,如监控
		eventboxReqContent := apistructs.GroupNotifyContent{
			SourceName:            "",
			SourceType:            "project",
			SourceID:              req.ProjectID,
			NotifyName:            notifyDetail.Name,
			NotifyItemDisplayName: notifyItem.DisplayName,
			Channels:              []apistructs.GroupNotifyChannel{},
			Label:                 notifyItem.Label,
			CalledShowNumber:      notifyItem.CalledShowNumber,
			OrgID:                 orgID,
		}

		err := e.bdl.CreateGroupNotifyEvent(apistructs.EventBoxGroupNotifyRequest{
			Sender:        "adapter",
			GroupID:       notifyDetail.NotifyGroup.ID,
			Channels:      notifyDetail.Channels,
			NotifyItem:    notifyItem,
			NotifyContent: &eventboxReqContent,
			Params:        params,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
