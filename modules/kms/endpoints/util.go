package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/kms/conf"
	"github.com/erda-project/erda/modules/kms/endpoints/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

// getPluginByKeyID 根据 keyID 获取对应的 plugin
func (e *Endpoints) getPluginByKeyID(keyID string) (kmstypes.Plugin, error) {
	store, err := e.KmsMgr.GetStore(conf.KmsStoreKind())
	if err != nil {
		return nil, err
	}
	keyInfo, err := store.GetKey(keyID)
	if err != nil {
		return nil, err
	}
	return e.KmsMgr.GetPlugin(keyInfo.GetPluginKind(), conf.KmsStoreKind())
}

// parseRequestBody return *errorresp.APIError
func (e *Endpoints) parseRequestBody(r *http.Request, req kmstypes.RequestValidator) *errorresp.APIError {
	if err := e.checkIdentity(r); err != nil {
		return apierrors.ErrCheckIdentity.InvalidParameter(err)
	}
	if r.ContentLength == 0 {
		return apierrors.ErrParseRequest.MissingParameter("request body")
	}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return apierrors.ErrParseRequest.InvalidParameter(err)
	}
	if err := req.ValidateRequest(); err != nil {
		return apierrors.ErrParseRequest.InvalidParameter(err)
	}
	return nil
}

func (e *Endpoints) checkIdentity(r *http.Request) (err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("check identity failed, err: %v", err)
			err = fmt.Errorf("invalid identity")
		}
	}()
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return err
	}
	if identityInfo.IsInternalClient() {
		return nil
	}
	return fmt.Errorf("not internal client")
}
