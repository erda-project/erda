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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// DeleteNotifySource 删除通知源
func (e *Endpoints) DeleteNotifySource(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotifySource.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	var deleteNotifySourceReq apistructs.DeleteNotifySourceRequest
	if err := json.NewDecoder(r.Body).Decode(&deleteNotifySourceReq); err != nil {
		return apierrors.ErrDeleteNotifySource.InvalidParameter("can't decode body").ToResp(), nil
	}
	if deleteNotifySourceReq.SourceID == "0" || deleteNotifySourceReq.SourceID == "" {
		return apierrors.ErrDeleteNotifySource.InvalidParameter("sourceId is null").ToResp(), nil
	}
	deleteNotifySourceReq.OrgID = orgID
	err = e.notifyGroup.DeleteNotifySource(&deleteNotifySourceReq)
	if err != nil {
		return apierrors.ErrDeleteNotifySource.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("")
}
