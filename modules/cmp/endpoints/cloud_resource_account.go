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
	"golang.org/x/text/message"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/impl/aliyun-resources/overview"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (e *Endpoints) CreateAccount(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	i18nPrinter := ctx.Value("i18nPrinter").(*message.Printer)
	orgid := r.Header.Get("Org-ID")
	req := apistructs.CreateCloudAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to decode CreateCloudAccountRequest: %v", err)
		return mkResponse(apistructs.CreateCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err := fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return errorresp.ErrResp(err)
	}

	// permission check
	err := e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.CreateAction)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	err = e.CloudAccount.Create(orgid, req.Vendor, req.AccessKey, req.Secret, req.Description)
	if err != nil {
		errstr := fmt.Sprintf("failed to create accountlist: %v, org: %s", err, orgid)
		return mkResponse(apistructs.CreateCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	// update cloud resource overview async when create account successfully
	go func() {
		ak_ctx, resp := e.mkCtx(ctx, orgid)
		if resp != nil {
			logrus.Infof("get ak_ctx failed, error:%s", resp.GetContent())
		}
		_, _ = overview.GetCloudResourceOverView(ak_ctx, i18nPrinter)
	}()

	return mkResponse(apistructs.CreateCloudAccountResponse{
		Header: apistructs.Header{Success: true},
	})
}

func (e *Endpoints) DeleteAccount(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgid := r.Header.Get("Org-ID")
	req := apistructs.DeleteCloudAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to decode DeleteCloudAccountRequest: %v", err)
		return mkResponse(apistructs.DeleteCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err := fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return errorresp.ErrResp(err)
	}

	// permission check
	err := e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.DeleteAction)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.CloudAccount.Delete(orgid, req.Vendor, req.AccessKey); err != nil {
		errstr := fmt.Sprintf("failed to delete account: %v, org: %s, ak: %s", err, orgid, req.AccessKey)
		return mkResponse(apistructs.DeleteCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.DeleteCloudAccountResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}

func (e *Endpoints) ListAccount(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgid := r.Header.Get("Org-ID")
	accounts, err := e.CloudAccount.List(orgid)
	if err != nil {
		errstr := fmt.Sprintf("failed to get accountlist: %v, org: %s", err, orgid)
		return mkResponse(apistructs.ListCloudAccountResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
			Data: apistructs.ListCloudAccountData{List: []apistructs.ListCloudAccount{}},
		})
	}

	result := []apistructs.ListCloudAccount{}
	for _, acc := range accounts {
		result = append(result, apistructs.ListCloudAccount{
			OrgID:       acc.Org,
			Vendor:      acc.Vendor,
			AccessKey:   acc.AccessKey,
			Description: acc.Description,
		})
	}
	return mkResponse(apistructs.ListCloudAccountResponse{
		Header: apistructs.Header{
			Success: true,
		},
		Data: apistructs.ListCloudAccountData{
			Total: len(result),
			List:  result,
		},
	})
}
