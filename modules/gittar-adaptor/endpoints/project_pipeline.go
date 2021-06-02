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
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (e *Endpoints) projectPipelineCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var createReq apistructs.PipelineCreateRequestV2
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		logrus.Errorf("[alert] failed to decode request body: %v", err)
		return apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("request body: %v", err)).ToResp(), nil
	}

	if createReq.Labels == nil {
		return apierrors.ErrCreatePipeline.InvalidParameter(fmt.Errorf("create req labels can not empty")).ToResp(), nil
	}

	if len(createReq.Labels[apistructs.LabelProjectID]) <= 0 {
		return apierrors.ErrCreatePipeline.InvalidParameter(fmt.Errorf("create req project label can not empty")).ToResp(), nil
	}

	projectID, err := strconv.ParseUint(createReq.Labels[apistructs.LabelProjectID], 10, 64)
	if err != nil {
		return nil, err
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  projectID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return apierrors.ErrCreatePipeline.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreatePipeline.AccessDenied().ToResp(), nil
		}
	}

	createReq.UserID = identityInfo.UserID

	spec, err := pipelineyml.New([]byte(createReq.PipelineYml))
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// add project labels and organization labels to snippet labels
	spec.Spec().LoopStagesActions(func(stage int, action *pipelineyml.Action) {
		if action.Type.IsSnippet() {
			if action.SnippetConfig == nil {
				err = fmt.Errorf("snippetConfig is empty")
				return
			}
			if action.SnippetConfig.Labels == nil {
				action.SnippetConfig.Labels = map[string]string{}
			}
			action.SnippetConfig.Labels[apistructs.LabelProjectID] = createReq.Labels[apistructs.LabelProjectID]
			action.SnippetConfig.Labels[apistructs.LabelOrgID] = createReq.Labels[apistructs.LabelOrgID]
		}
	})
	if err != nil {
		return errorresp.ErrResp(err)
	}

	yml, err := pipelineyml.GenerateYml(spec.Spec())
	if err != nil {
		return errorresp.ErrResp(err)
	}
	createReq.PipelineYml = string(yml)

	resp, err := e.pipeline.CreatePipelineV2(&createReq)
	if err != nil {
		logrus.Errorf("create pipeline failed, reqPipeline: %+v, (%+v)", createReq, err)
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(resp)
}
