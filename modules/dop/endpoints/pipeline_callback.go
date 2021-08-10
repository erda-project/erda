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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// CDPCallback cdp hook的回调
func (e *Endpoints) CDPCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req           apistructs.PipelineInstanceEvent
		runningTaskID int64
	)
	if r.Body == nil {
		return apierrors.ErrDealCDPCallback.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDealCDPCallback.InvalidParameter(err).ToResp(), nil
	}

	go func() {
		err := e.cdp.CdpNotifyProcess(&req)
		if err != nil {
			logrus.Errorf("failed to process cdp notify %s", err)
		}
	}()
	return httpserver.OkResp(runningTaskID)
}
