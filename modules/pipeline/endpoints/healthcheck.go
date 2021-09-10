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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) healthCheck(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	_, err := e.dbClient.Exec("select 1")
	if err != nil {
		logrus.Errorf("failed to health check, err: %v", err)

		return apierrors.ErrPipelineHealthCheck.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("success")
}
