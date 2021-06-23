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

// WatchApprovalStatusChanged 监听审批流状态变更，同步审批流状态至依赖方
func (e *Endpoints) WatchApprovalStatusChanged(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		event apistructs.ApprovalStatusChangedEvent
		err   error
	)
	if err = json.NewDecoder(r.Body).Decode(&event); err != nil {
		return apierrors.ErrApprovalStatusChanged.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("approvalStatusChangedEvent: %+v", event)

	// 处理审批流状态变更通知
	switch event.Content.ApprovalType {
	case apistructs.ApproveCeritficate:
		err = e.appCertificate.ModifyApprovalStatus(int64(event.Content.ApprovalID), event.Content.ApprovalStatus)
	case apistructs.ApproveLibReference:
		err = e.libReference.UpdateApprovalStatus(event.Content.ApprovalID, event.Content.ApprovalStatus)
	}

	if err != nil {
		return apierrors.ErrApprovalStatusChanged.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("handle success")
}
