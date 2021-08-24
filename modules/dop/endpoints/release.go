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
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/http/httpserver"
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
	rules, err := e.branchRule.Query(apistructs.AppScope, int64(appID))
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
