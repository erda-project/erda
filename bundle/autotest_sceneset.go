package bundle

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetSceneSets(req apistructs.SceneSetRequest) ([]apistructs.SceneSet, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}

	request := b.hc.Get(host).Path("/api/autotests/scenesets")
	var rsp apistructs.GetSceneSetsResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		Params(req.URLQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	return rsp.Data, nil
}

func (b *Bundle) GetSceneSet(req apistructs.SceneSetRequest) (*apistructs.SceneSet, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request := b.hc.Get(host).Path("/api/autotests/scenesets/" + strconv.FormatInt(int64(req.SetID), 10))
	var rsp apistructs.GetSceneSetResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	return &rsp.Data, nil
}

func (b *Bundle) CreateSceneSet(req apistructs.SceneSetRequest) (*uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}

	request := b.hc.Post(host).Path("/api/autotests/scenesets")
	var rsp apistructs.CreateSceneSetResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	return &rsp.Id, nil
}

func (b *Bundle) UpdateSceneSet(req apistructs.SceneSetRequest) (*apistructs.SceneSet, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}

	request := b.hc.Put(host).Path("/api/autotests/scenesets/" + strconv.FormatInt(int64(req.SetID), 10))
	var rsp apistructs.UpdateSceneSetResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	return &rsp.Data, nil
}

func (b *Bundle) DeleteSceneSet(req apistructs.SceneSetRequest) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}

	request := b.hc.Delete(host).Path("/api/autotests/scenesets/" + strconv.FormatInt(int64(req.SetID), 10))
	var rsp apistructs.DeleteSceneSetResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return toAPIError(resp.StatusCode(), rsp.Error)
	}
	return nil
}

func (b *Bundle) DragSceneSet(req apistructs.SceneSetRequest) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}

	request := b.hc.Put(host).Path("/api/autotests/scenesets/actions/drag")
	var rsp apistructs.DeleteSceneSetResponse
	resp, err := request.
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().JSON(&rsp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return toAPIError(resp.StatusCode(), rsp.Error)
	}
	return nil
}
