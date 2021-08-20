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

	libstr "github.com/appscode/go/strings"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

// Get Addon config
func (e *Endpoints) GetAddonConfig(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	addonID := r.URL.Query().Get("addonID")

	if libstr.IsEmpty(&addonID) {
		return apierrors.ErrGetAddonConfig.InvalidParameter("addonID").ToResp(), nil
	}

	_, resp := e.GetIdentity(r)
	if resp != nil {
		return resp, nil
	}

	result, err := e.Addons.GetAddonConfig(addonID)
	if err != nil {
		logrus.Errorf("get addon config failed, addonID:%s, error:%v", addonID, err)
		return apierrors.ErrGetAddonConfig.InternalError(err).ToResp(), nil
	}

	return mkResponse(apistructs.AddonConfigResponse{
		Header: apistructs.Header{Success: true},
		Data:   result,
	})
}

// Get Addon status
func (e *Endpoints) GetAddonStatus(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	addonID := r.URL.Query().Get("addonID")
	addonName := r.URL.Query().Get("addonName")

	if libstr.IsEmpty(&addonID) || libstr.IsEmpty(&addonName) {
		return apierrors.ErrGetAddonConfig.InvalidParameter("addonID or addonName is empty").ToResp(), nil
	}

	_, resp := e.GetIdentity(r)
	if resp != nil {
		return resp, nil
	}

	status, err := e.Addons.GetAddonStatus(addonName, addonID)
	if err != nil {
		logrus.Errorf("get addon status failed, %v", err)
		return mkResponse(apistructs.OpsAddonStatusResponse{
			Header: apistructs.Header{Success: true},
			Data:   apistructs.OpsAddonStatusData{Status: apistructs.StatusUnknown},
		})
	}

	return mkResponse(apistructs.OpsAddonStatusResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.OpsAddonStatusData{Status: status},
	})
}

// Update addon config
func (e *Endpoints) UpdateAddonConfig(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	req := apistructs.AddonConfigUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err := fmt.Errorf("failed to decode UpdateAddonConfig: %v", err)
		logrus.Errorf(err.Error())
		return mkResponse(apistructs.CreateCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err := fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return errorresp.ErrResp(err)
	}

	// permission check
	err := e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.UpdateAction)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	err = e.Addons.UpdateAddonConfig(req)

	if err != nil {
		logrus.Errorf("update addon config failed, request:%+v, error:%v", req, err)
		return apierrors.ErrUpdateAddonConfig.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

// Addon Scale
func (e *Endpoints) AddonScale(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	req := apistructs.AddonScaleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err := fmt.Errorf("failed to decode UpdateAddonConfig: %v", err)
		logrus.Errorf(err.Error())
		return mkResponse(apistructs.CreateCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err := fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return errorresp.ErrResp(err)
	}

	// permission check
	err := e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.UpdateAction)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	err = e.Addons.AddonScale(i, req)
	if err != nil {
		logrus.Errorf("addon scale failed, request:%+v, error:%v", req, err)
		return apierrors.ErrUpdateAddonConfig.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}
