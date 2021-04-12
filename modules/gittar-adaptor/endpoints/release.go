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

// Package endpoints release的handle
package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// ReleaseCallback gittar hook的回调，目前只响应TAG，自动触发pipeline进行release
func (e *Endpoints) ReleaseCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.GittarPushEvent

	if r.Body == nil {
		logrus.Errorf("nil body")
		return apierrors.ErrReleaseCallback.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrReleaseCallback.InvalidParameter(err).ToResp(), nil
	}
	logrus.Debugf("GittarCallback: request body: %+v", req)
	params := map[string]string{
		"beforeID":     req.Content.Before,
		"afterID":      req.Content.After,
		"ref":          req.Content.Ref,
		"pusherName":   req.Content.Pusher.NickName,
		"commitsCount": strconv.FormatInt(int64(req.Content.TotalCommitsCount), 10),
	}
	e.TriggerGitNotify(req.OrgID, req.ApplicationID, apistructs.GitPushEvent, params)

	appID, _ := strconv.ParseUint(req.ApplicationID, 10, 64)
	rules, err := e.bdl.GetAppBranchRules(appID)
	if err != nil {
		logrus.Errorf("failed to get branch rules, req: %+v, (%+v)", req, err)
		return apierrors.ErrReleaseCallback.InternalError(err).ToResp(), nil
	}
	refName := ""
	if req.Content.IsTag {
		refName = strings.TrimPrefix(req.Content.Ref, "refs/tags/")
	} else {
		refName = strings.TrimPrefix(req.Content.Ref, "refs/heads/")
	}

	// 从 gittar 获取 pipeline.yml
	strPipelineYml, err := e.pipeline.FetchPipelineYml(req.Content.Repository.URL, refName, apistructs.DefaultPipelineYmlName)
	if err != nil {
		logrus.Errorf("failed to fetch pipeline.yml from gittar, req: %+v, (%+v)", req, err)
		return apierrors.ErrReleaseCallback.InternalError(err).ToResp(), nil
	}

	pipelineYml, err := pipelineyml.New([]byte(strPipelineYml))
	if err != nil {
		logrus.Errorf("failed to parse pipeline.yml yaml:%v \n err:%v", pipelineYml, err)
		return apierrors.ErrReleaseCallback.InternalError(err).ToResp(), nil
	}

	// 应用级设置
	if pipelineYml.Spec().On != nil && pipelineYml.Spec().On.Push != nil {
		if !diceworkspace.IsRefPatternMatch(refName, pipelineYml.Spec().On.Push.Branches) {
			return httpserver.OkResp("")
		}
	} else {
		// 项目级设置
		validBranch := diceworkspace.GetValidBranchByGitReference(refName, rules)
		if !validBranch.IsTriggerPipeline {
			return httpserver.OkResp("")
		}
	}

	// 创建pipeline流程
	reqPipeline := &apistructs.PipelineCreateRequest{
		AppID:              uint64(req.Content.Repository.ApplicationID),
		Branch:             refName,
		Source:             apistructs.PipelineSourceDice,
		PipelineYmlSource:  apistructs.PipelineYmlSourceGittar,
		PipelineYmlContent: strPipelineYml,
		AutoRun:            true,
		UserID:             req.Content.Pusher.ID,
	}

	v2, err := e.pipeline.ConvertPipelineToV2(reqPipeline)
	if err != nil {
		logrus.Errorf("error convert to pipelineV2 %s, (%+v)", strPipelineYml, err)
		return apierrors.ErrReleaseCallback.InternalError(err).ToResp(), nil
	}
	v2.ForceRun = true

	resp, err := e.pipeline.CreatePipelineV2(v2)
	if err != nil {
		logrus.Errorf("create pipeline failed, pipeline: %s, (%+v)", strPipelineYml, err)
		return apierrors.ErrReleaseCallback.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp)
}
