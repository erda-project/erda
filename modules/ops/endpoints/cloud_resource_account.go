package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/impl/aliyun-resources/overview"
	"github.com/erda-project/erda/pkg/httpserver"
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

	err := e.CloudAccount.Create(orgid, req.Vendor, req.AccessKey, req.Secret, req.Description)
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
