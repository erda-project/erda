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

package apidocsvc

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/apim/services/websocket"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// http ==> ws
// 直接将错误处理写到了 w http.ResponseWriter, 所以调用方不必对错误进行额外处理了
func (svc *Service) Upgrade(w http.ResponseWriter, r *http.Request, req *apistructs.WsAPIDocHandShakeReq) *errorresp.APIError {
	ft, err := bundle.NewGittarFileTree(req.URIParams.Inode)
	if err != nil {
		return apierrors.WsUpgrade.InvalidParameter("不合法的 inode")
	}
	appID, err := strconv.ParseUint(ft.ApplicationID(), 10, 64)
	if err != nil {
		return apierrors.WsUpgrade.InvalidParameter("不合法的 inode")
	}

	h := APIDocWSHandler{
		orgID:     req.OrgID,
		userID:    req.Identity.UserID,
		appID:     appID,
		branch:    ft.BranchName(),
		filename:  filepath.Base(ft.PathFromRepoRoot()),
		req:       req,
		svc:       svc,
		sessionID: uuid.New().String(),
		ft:        ft,
	}

	ws := websocket.New()
	ws.Register(heartBeatRequest, h.wrap(h.handleHeartBeat))
	ws.Register(autoSaveRequest, h.wrap(h.handleAutoSave))
	ws.Register(commitRequest, h.wrap(h.handleCommit))
	ws.AfterConnected(h.AfterConnected)
	ws.BeforeClose(h.BeforeClose)
	if err := ws.Upgrade(w, r); err != nil {
		return apierrors.WsUpgrade.InvalidParameter(err)
	}
	ws.Run()

	return nil
}
