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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// gittar yml change update pipeline definition
func (e *Endpoints) gittarPipelineDefinitionUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.GittarPushPayloadEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("gittarPipelineDefinitionUpdate decode error %v", err)
		return apierrors.ErrUpdatePipelineDefinition.InvalidParameter(err).ToResp(), nil
	}

	if err := e.pipeline.PipelineDefinitionUpdate(req); err != nil {
		logrus.Errorf("gittarPipelineDefinitionUpdate pipelineDefinitionUpdate error %v", err)
		return apierrors.ErrUpdatePipelineDefinition.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}
