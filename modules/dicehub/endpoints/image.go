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
	"net/url"

	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetImage 获取镜像
// TODO 参数校验优化，增加异常场景处理
func (e *Endpoints) GetImage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrGetImage.NotLogin().ToResp(), nil
	}

	imageIDOrImage, err := url.QueryUnescape(vars["imageIdOrImage"])
	if err != nil {
		return apierrors.ErrGetImage.InvalidParameter("imageId").ToResp(), nil
	}

	image, err := e.image.Get(imageIDOrImage)
	if err != nil {
		return apierrors.ErrGetImage.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(image)
}

// ListImage 镜像列表
func (e *Endpoints) ListImage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrListImage.NotLogin().ToResp(), nil
	}

	pageNum := r.URL.Query().Get("pageNum")
	if pageNum == "" {
		pageNum = "1"
	}
	num, err := strutil.Atoi64(pageNum)
	if err != nil {
		return apierrors.ErrListImage.InvalidParameter("pageNum").ToResp(), nil
	}

	pageSize := r.URL.Query().Get("pageSize")
	if pageSize == "" {
		pageSize = "20"
	}
	size, err := strutil.Atoi64(pageSize)
	if err != nil {
		return apierrors.ErrListImage.InvalidParameter("pageSize").ToResp(), nil
	}

	resp, err := e.image.List(orgID, num, size)
	if err != nil {
		return apierrors.ErrListImage.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp)
}
