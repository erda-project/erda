package endpoints

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) createCmsNs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreatePipelineCmsNs.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrCreatePipelineCmsNs.AccessDenied().ToResp(), nil
	}

	// 参数解析
	var req apistructs.PipelineCmsCreateNsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreatePipelineCmsNs.InvalidParameter(err).ToResp(), nil
	}

	if err := e.cmSvc.CreateNS(req.PipelineSource, req.NS); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) updateCmsNsConfigs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdatePipelineCmsConfigs.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrUpdatePipelineCmsConfigs.AccessDenied().ToResp(), nil
	}

	// 参数解析
	// ns
	ns := vars[pathNs]
	bodyByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apierrors.ErrUpdatePipelineCmsConfigs.InvalidParameter(errors.Errorf("read request body err: %v", err.Error())).ToResp(), nil
	}
	// req
	var req apistructs.PipelineCmsUpdateConfigsRequest
	if err := json.Unmarshal(bodyByte, &req); err != nil {
		var reqV1 apistructs.PipelineCmsUpdateConfigsRequestV1
		if err := json.Unmarshal(bodyByte, &reqV1); err != nil {
			return apierrors.ErrUpdatePipelineCmsConfigs.InvalidParameter(err).ToResp(), nil
		}
		// transform reqV1 to req
		req.KVs = make(map[string]apistructs.PipelineCmsConfigValue, len(reqV1.KVs))
		for k, v := range reqV1.KVs {
			req.KVs[k] = apistructs.PipelineCmsConfigValue{
				Value:       v,
				EncryptInDB: true,
				Type:        apistructs.PipelineCmsConfigTypeKV,
				Operations: &apistructs.PipelineCmsConfigOperations{
					CanDownload: false,
					CanEdit:     true,
					CanDelete:   true,
				},
			}
		}
		// 兼容 cdp
		nsPrefix := strutil.TrimPrefixes(ns, "cdp-action-")
		switch nsPrefix {
		case "dev":
			req.PipelineSource = apistructs.PipelineSourceCDPDev
		case "test":
			req.PipelineSource = apistructs.PipelineSourceCDPTest
		case "staging":
			req.PipelineSource = apistructs.PipelineSourceCDPStaging
		case "prod":
			req.PipelineSource = apistructs.PipelineSourceCDPProd
		default:
			req.PipelineSource = apistructs.PipelineSourceDefault
		}
	}

	if err := e.cmSvc.UpdateConfigs(req.PipelineSource, ns, req.KVs); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) deleteCmsNsConfigs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdatePipelineCmsConfigs.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrUpdatePipelineCmsConfigs.AccessDenied().ToResp(), nil
	}

	// 参数解析
	// ns
	ns := vars[pathNs]
	// req
	var req apistructs.PipelineCmsDeleteConfigsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDeletePipelineCmsConfigs.InvalidParameter(err).ToResp(), nil
	}

	var opErr error
	// 删除 ns
	if req.DeleteNS {
		opErr = e.cmSvc.DeleteNS(req.PipelineSource, ns)
	} else {
		opErr = e.cmSvc.DeleteConfigs(req.PipelineSource, ns, req.DeleteKeys, req.DeleteForce)
	}

	if opErr != nil {
		return errorresp.ErrResp(opErr)
	}
	return httpserver.OkResp(nil)
}

func (e *Endpoints) getCmsNsConfigs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetPipelineCmsConfigs.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrGetPipelineCmsConfigs.AccessDenied().ToResp(), nil
	}

	// 参数解析
	// ns
	ns := vars[pathNs]
	// keys
	var req apistructs.PipelineCmsGetConfigsRequest
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return apierrors.ErrGetPipelineCmsConfigs.InvalidParameter(err).ToResp(), nil
		}
	}

	kvs, err := e.cmSvc.GetConfigs(req.PipelineSource, ns, req.GlobalDecrypt, req.Keys...)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(kvs)
}

func (e *Endpoints) listCmsNs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListPipelineCmsNs.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrListPipelineCmsNs.AccessDenied().ToResp(), nil
	}

	// 参数解析
	var req apistructs.PipelineCmsListNsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrListPipelineCmsNs.InvalidParameter(err).ToResp(), nil
	}

	namespaces, err := e.cmSvc.PrefixListNS(req.PipelineSource, req.NsPrefix)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(namespaces)
}
