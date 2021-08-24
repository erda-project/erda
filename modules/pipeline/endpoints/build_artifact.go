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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

// queryBuildArtifact 用于外部用户查询
func (e *Endpoints) queryBuildArtifact(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	sha := vars[pathSha]
	artifact, err := e.buildArtifactSvc.Query(sha)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(artifact.Convert2DTO())
}

// registerBuildArtifact 用于外部用户主动注册 artifact
func (e *Endpoints) registerBuildArtifact(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.BuildArtifactRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrRegisterBuildArtifact.InvalidParameter(err).ToResp(), nil
	}

	artifact, err := e.buildArtifactSvc.Register(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(artifact.Convert2DTO())
}
