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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetAutoTestSceneStep(req apistructs.AutotestGetSceneStepReq) (*apistructs.AutoTestSceneStep, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.AutotestGetSceneStepResp
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/autotests/scenes-step/"+strconv.FormatInt(int64(req.ID), 10))).
		Header(httputil.UserHeader, req.UserID).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp.Data, nil
}

func (b *Bundle) ListAutoTestSceneStep(req apistructs.AutotestSceneRequest) ([]apistructs.AutoTestSceneStep, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.AutotestGetSceneStepResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/autotests/scenes/"+strconv.FormatInt(int64(req.SceneID), 10)+"/actions/get-step")).
		Header(httputil.UserHeader, req.UserID).
		Params(req.URLQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

func (b *Bundle) UpdateAutoTestSceneStep(req apistructs.AutotestSceneRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc
	var rsp apistructs.AutotestCreateSceneResponse
	httpResp, err := hc.Put(host).Path(fmt.Sprintf("/api/autotests/scenes-step/"+strconv.FormatInt(int64(req.ID), 10))).
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return 0, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

func (b *Bundle) MoveAutoTestSceneStep(req apistructs.AutotestSceneRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc
	var rsp apistructs.AutotestCreateSceneResponse
	httpResp, err := hc.Put(host).Path(fmt.Sprintf("/api/autotests/scenes-step/actions/move")).
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return 0, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

func (b *Bundle) DeleteAutoTestSceneStep(req apistructs.AutotestSceneRequest) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}
	hc := b.hc
	var rsp apistructs.AutotestCancelSceneResponse
	httpResp, err := hc.Delete(host).Path(fmt.Sprintf("/api/autotests/scenes-step/"+strconv.FormatInt(int64(req.ID), 10))).
		JSONBody(&req).
		Header(httputil.UserHeader, req.UserID).
		Do().JSON(&rsp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return nil
}

func (b *Bundle) CreateAutoTestSceneStep(req apistructs.AutotestSceneRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc
	var rsp apistructs.AutotestCreateSceneResponse
	httpResp, err := hc.Post(host).Path(fmt.Sprintf("/api/autotests/scenes/"+strconv.FormatInt(int64(req.SceneID), 10)+"/actions/add-step")).
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return 0, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}
