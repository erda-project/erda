package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/modules/kms/conf"
	"github.com/erda-project/erda/modules/kms/endpoints/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

func (e *Endpoints) KmsCreateKey(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.CreateKeyRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.KmsMgr.GetPlugin(req.PluginKind, conf.KmsStoreKind())
	if err != nil {
		return apierrors.ErrCreateKey.InternalError(err).ToResp(), nil
	}

	createKeyResp, err := plugin.CreateKey(ctx, &req)
	if err != nil {
		return apierrors.ErrCreateKey.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(createKeyResp)
}

func (e *Endpoints) DescribeKey(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.DescribeKeyRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.getPluginByKeyID(req.KeyID)
	if err != nil {
		return apierrors.ErrDescribeKey.InternalError(err).ToResp(), nil
	}

	descResp, err := plugin.DescribeKey(ctx, &req)
	if err != nil {
		return apierrors.ErrDescribeKey.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(descResp)
}
