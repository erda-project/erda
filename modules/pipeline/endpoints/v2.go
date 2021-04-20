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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) pipelineCreateV2(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var createReq apistructs.PipelineCreateRequestV2
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		logrus.Errorf("[alert] failed to decode request body: %v", err)
		return apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("request body: %v", err)).ToResp(), nil
	}
	// 兼容处理
	if createReq.AutoRun {
		createReq.AutoRunAtOnce = true
	}

	logrus.Debugf("pipeline create v2 request: %+v\n", createReq)

	// 身份校验
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	createReq.IdentityInfo = identityInfo

	p, err := e.pipelineSvc.CreateV2(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(e.pipelineSvc.ConvertPipeline(p))
}
