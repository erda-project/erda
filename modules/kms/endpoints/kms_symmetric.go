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

	"github.com/erda-project/erda/modules/kms/endpoints/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

func (e *Endpoints) KmsEncrypt(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.EncryptRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.getPluginByKeyID(req.KeyID)
	if err != nil {
		return apierrors.ErrEncrypt.InternalError(err).ToResp(), nil
	}
	encryptResp, err := plugin.Encrypt(ctx, &req)
	if err != nil {
		return apierrors.ErrEncrypt.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(encryptResp)
}

func (e *Endpoints) KmsDecrypt(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.DecryptRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.getPluginByKeyID(req.KeyID)
	if err != nil {
		return apierrors.ErrDecrypt.InternalError(err).ToResp(), nil
	}
	decryptResp, err := plugin.Decrypt(ctx, &req)
	if err != nil {
		return apierrors.ErrDecrypt.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(decryptResp)
}

func (e *Endpoints) KmsGenerateDataKey(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.GenerateDataKeyRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.getPluginByKeyID(req.KeyID)
	if err != nil {
		return apierrors.ErrGenerateDataKey.InvalidParameter(err).ToResp(), nil
	}
	generateResp, err := plugin.GenerateDataKey(ctx, &req)
	if err != nil {
		return apierrors.ErrGenerateDataKey.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(generateResp)
}

func (e *Endpoints) KmsRotateKeyVersion(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req kmstypes.RotateKeyVersionRequest
	if err := e.parseRequestBody(r, &req); err != nil {
		return err.ToResp(), nil
	}

	plugin, err := e.getPluginByKeyID(req.KeyID)
	if err != nil {
		return apierrors.ErrRotateKeyVersion.InvalidParameter(err).ToResp(), nil
	}
	rotateResp, err := plugin.RotateKeyVersion(ctx, &req)
	if err != nil {
		return apierrors.ErrRotateKeyVersion.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(rotateResp)
}
