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

package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) CreatePipelineCmsNs(req apistructs.PipelineCmsCreateNsRequest) error {
	host, err := b.urls.Pipeline()
	if err != nil {
		return err
	}
	hc := b.hc

	var createResp apistructs.PipelineCmsCreateNsResponse
	httpResp, err := hc.Post(host).Path("/api/pipelines/cms/ns").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&createResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !createResp.Success {
		return toAPIError(httpResp.StatusCode(), createResp.Error)
	}
	return nil
}

func (b *Bundle) CreateOrUpdatePipelineCmsNsConfigs(ns string, req apistructs.PipelineCmsUpdateConfigsRequest) error {
	host, err := b.urls.Pipeline()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.PipelineCmsUpdateConfigsResponse
	httpResp, err := hc.Post(host).Path(fmt.Sprintf("/api/pipelines/cms/ns/%s", ns)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !updateResp.Success {
		return toAPIError(httpResp.StatusCode(), updateResp.Error)
	}
	return nil
}

func (b *Bundle) DeletePipelineCmsNsConfigs(ns string, req apistructs.PipelineCmsDeleteConfigsRequest) error {
	host, err := b.urls.Pipeline()
	if err != nil {
		return err
	}
	hc := b.hc

	var delResp apistructs.PipelineCmsDeleteConfigsResponse
	httpResp, err := hc.Delete(host).Path(fmt.Sprintf("/api/pipelines/cms/ns/%s", ns)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&delResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !delResp.Success {
		return toAPIError(httpResp.StatusCode(), delResp.Error)
	}
	return nil
}

func (b *Bundle) GetPipelineCmsNsConfigs(ns string, req apistructs.PipelineCmsGetConfigsRequest) ([]apistructs.PipelineCmsConfig, error) {
	host, err := b.urls.Pipeline()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.PipelineCmsGetConfigsResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/pipelines/cms/ns/%s", ns)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !getResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}

func (b *Bundle) ListPipelineCmsNs(req apistructs.PipelineCmsListNsRequest) ([]apistructs.PipelineCmsNs, error) {
	host, err := b.urls.Pipeline()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var listResp apistructs.PipelineCmsListNsResponse
	httpResp, err := hc.Get(host).Path("/api/pipelines/cms/ns").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !listResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), listResp.Error)
	}
	return listResp.Data, nil
}
