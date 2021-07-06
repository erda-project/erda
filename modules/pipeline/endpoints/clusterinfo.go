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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) clusterHook(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	req := apistructs.ClusterEvent{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errStr := fmt.Sprintf("failed to decode clusterhook request, err: %v", err)
		logrus.Error(errStr)
		return httpserver.ErrResp(http.StatusBadRequest, "", errStr)
	}
	if err := e.pipelineSvc.ClusterHook(req); err != nil {
		errStr := fmt.Sprintf("failed to handle cluster event, err: %v", err)
		logrus.Error(errStr)
		return httpserver.ErrResp(http.StatusBadRequest, "", errStr)
	}
	return httpserver.OkResp(nil)
}
