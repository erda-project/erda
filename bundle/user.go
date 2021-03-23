package bundle

import (
	"net/url"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetCurrentUser(userID string) (*apistructs.UserInfo, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var userResp apistructs.UserCurrentResponse
	resp, err := hc.Get(host).Path("/api/users/current").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.UserHeader, userID).
		Do().JSON(&userResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !userResp.Success {
		return nil, toAPIError(resp.StatusCode(), userResp.Error)
	}
	return &userResp.Data, nil
}

func (b *Bundle) ListUsers(req apistructs.UserListRequest) (*apistructs.UserListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var userResp apistructs.UserListResponse
	resp, err := hc.Get(host).Path("/api/users").
		Header(httputil.InternalHeader, "bundle").
		Param("q", req.Query).
		Param("plaintext", strconv.FormatBool(req.Plaintext)).
		Params(url.Values{"userID": req.UserIDs}).
		Do().JSON(&userResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !userResp.Success {
		return nil, toAPIError(resp.StatusCode(), userResp.Error)
	}
	return &userResp.Data, nil
}
