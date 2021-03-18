package bundle

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) ListAutoTestSceneOutput(req apistructs.AutotestSceneRequest) ([]apistructs.AutoTestSceneOutput, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.AutotestGetSceneOutputResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/autotests/scenes/"+strconv.FormatInt(int64(req.SceneID), 10)+"/actions/list-output")).
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

func (b *Bundle) ListAutoTestStepOutput(req apistructs.AutotestListStepOutPutRequest) (map[string]string, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.AutotestGetSceneStepOutPutResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/autotests/scenes-step-output")).
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

func (b *Bundle) UpdateAutoTestSceneOutput(req apistructs.AutotestSceneOutputUpdateRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc
	var rsp apistructs.AutotestCreateSceneResponse
	httpResp, err := hc.Put(host).Path(fmt.Sprintf("/api/autotests/scenes/"+strconv.FormatInt(int64(req.SceneID), 10)+"/actions/update-output")).
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

func (b *Bundle) CreateAutoTestSceneOutput(req apistructs.AutotestSceneRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc
	var rsp apistructs.AutotestCreateSceneResponse
	httpResp, err := hc.Post(host).Path(fmt.Sprintf("/api/autotests/scenes/"+strconv.FormatInt(int64(req.SceneID), 10)+"/actions/add-output")).
		Header(httputil.UserHeader, req.UserID).
		Params(req.URLQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return 0, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}
