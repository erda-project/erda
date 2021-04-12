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

// RepoMrEventCallback Mr事件回调
func (e *Endpoints) RepoMrEventCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var gitEvent apistructs.RepoCreateMrEvent

	if r.Body == nil {
		logrus.Errorf("nil body")
		return apierrors.ErrRepoMrCallback.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&gitEvent); err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrRepoMrCallback.InvalidParameter(err).ToResp(), nil
	}

	params := map[string]string{
		"mr_title":     gitEvent.Content.Title,
		"sourceBranch": gitEvent.Content.SourceBranch,
		"targetBranch": gitEvent.Content.TargetBranch,
		"authorName":   gitEvent.Content.AuthorUser.NickName,
		"assigneeName": gitEvent.Content.AssigneeUser.NickName,
		"state":        gitEvent.Content.State,
		"link":         gitEvent.Content.Link,
		"atUserIDs":    gitEvent.Content.AssigneeId,
		"description":  gitEvent.Content.Description,
	}
	switch gitEvent.Content.EventName {
	case apistructs.GitCommentMREvent, apistructs.GitMergeMREvent, apistructs.GitCloseMREvent:
		params["atUserIDs"] = gitEvent.Content.AuthorId
	}

	e.TriggerGitNotify(gitEvent.OrgID, gitEvent.ApplicationID, gitEvent.Content.EventName, params)
	return httpserver.OkResp("")
}

// RepoBranchEventCallback Branch事件回调
func (e *Endpoints) RepoBranchEventCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var gitEvent apistructs.RepoBranchEvent

	if r.Body == nil {
		logrus.Errorf("nil body")
		return apierrors.ErrRepoMrCallback.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&gitEvent); err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrRepoMrCallback.InvalidParameter(err).ToResp(), nil
	}

	params := map[string]string{
		"branch":       gitEvent.Content.Name,
		"operatorName": gitEvent.Content.OperatorName,
		"link":         gitEvent.Content.Link,
	}
	e.TriggerGitNotify(gitEvent.OrgID, gitEvent.ApplicationID, gitEvent.Content.EventName, params)
	return httpserver.OkResp("")
}

// RepoTagEventCallback Tag事件回调
func (e *Endpoints) RepoTagEventCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var gitEvent apistructs.RepoTagEvent

	if r.Body == nil {
		logrus.Errorf("nil body")
		return apierrors.ErrRepoMrCallback.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&gitEvent); err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrRepoMrCallback.InvalidParameter(err).ToResp(), nil
	}

	params := map[string]string{
		"tag":          gitEvent.Content.Name,
		"operatorName": gitEvent.Content.OperatorName,
		"link":         gitEvent.Content.Link,
		"commitId":     gitEvent.Content.Object,
		"message":      gitEvent.Content.Message,
	}
	e.TriggerGitNotify(gitEvent.OrgID, gitEvent.ApplicationID, gitEvent.Content.EventName, params)
	return httpserver.OkResp("")
}

func (e *Endpoints) TriggerGitNotify(orgID string, appID string, eventName string, params map[string]string) error {
	sourceType := "app"
	notifyDetails, err := e.bdl.QueryNotifiesBySource(orgID, sourceType, appID, eventName, "")
	if err != nil {
		logrus.Errorf("failed to query notifies by source err:%s", err)
		return err
	}

	for _, notifyDetail := range notifyDetails {
		if notifyDetail.NotifyGroup == nil {
			continue
		}
		if len(notifyDetail.NotifyItems) == 0 {
			continue
		}
		notifyItem := notifyDetail.NotifyItems[0]
		orgID, _ := strconv.ParseInt(orgID, 10, 64)
		appId, _ := strconv.ParseInt(appID, 10, 64)
		appDTO, err := e.bdl.GetApp(uint64(appId))
		if err != nil {
			logrus.Errorf("failed to get app info %s", err)
		}
		eventboxReqContent := apistructs.GroupNotifyContent{
			SourceName:            sourceType + "-" + appID,
			SourceType:            sourceType,
			SourceID:              appID,
			NotifyName:            notifyDetail.Name,
			NotifyItemDisplayName: notifyItem.DisplayName,
			Channels:              []apistructs.GroupNotifyChannel{},
			Label:                 notifyItem.Label,
			CalledShowNumber:      notifyItem.CalledShowNumber,
			OrgID:                 orgID,
		}
		params["appName"] = appDTO.Name
		params["projectName"] = appDTO.ProjectName

		err = e.bdl.CreateGroupNotifyEvent(apistructs.EventBoxGroupNotifyRequest{
			Sender:        "adapter",
			GroupID:       notifyDetail.NotifyGroup.ID,
			Channels:      notifyDetail.Channels,
			NotifyItem:    notifyItem,
			NotifyContent: &eventboxReqContent,
			Params:        params,
		})
		if err != nil {
			logrus.Errorf("failed to create group notify event %s", err.Error())
		}
	}
	return nil
}
