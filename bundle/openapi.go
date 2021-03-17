package bundle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// applyOpenAPIToken 从 openapi 动态获取 oauth2 token
func (b *Bundle) GetOpenapiOAuth2Token(req apistructs.OpenapiOAuth2TokenGetRequest) (*apistructs.OpenapiOAuth2Token, error) {
	host, err := b.urls.Openapi()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var body bytes.Buffer
	httpResp, err := hc.Post(host).Path(fmt.Sprintf("/oauth2/token")).
		Header(httputil.InternalHeader, "bundle").
		Param("grant_type", "client_credentials").
		Param("client_id", req.ClientID).
		Param("client_secret", req.ClientSecret).
		JSONBody(&req.Payload).
		Do().Body(&body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(errors.New(body.String()))
	}
	var token apistructs.OpenapiOAuth2Token
	err = json.NewDecoder(&body).Decode(&token)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &token, nil
}

func (b *Bundle) InvalidateOpenapiOAuth2Token(req apistructs.OpenapiOAuth2TokenInvalidateRequest) (*apistructs.OpenapiOAuth2Token, error) {
	host, err := b.urls.Openapi()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var body bytes.Buffer
	httpResp, err := hc.Post(host).Path(fmt.Sprintf("/oauth2/invalidate_token")).
		Header(httputil.InternalHeader, "bundle").
		Param("access_token", req.AccessToken).
		Do().Body(&body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(errors.New(body.String()))
	}
	var token apistructs.OpenapiOAuth2Token
	err = json.NewDecoder(&body).Decode(&token)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &token, nil
}
