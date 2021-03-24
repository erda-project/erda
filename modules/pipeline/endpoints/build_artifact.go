package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// queryBuildArtifact 用于外部用户查询
func (e *Endpoints) queryBuildArtifact(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	sha := vars[pathSha]
	artifact, err := e.buildArtifactSvc.Query(sha)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(artifact.Convert2DTO())
}

// registerBuildArtifact 用于外部用户主动注册 artifact
func (e *Endpoints) registerBuildArtifact(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.BuildArtifactRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrRegisterBuildArtifact.InvalidParameter(err).ToResp(), nil
	}

	artifact, err := e.buildArtifactSvc.Register(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(artifact.Convert2DTO())
}
