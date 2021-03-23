package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) reloadActionExecutorConfig(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	_, cfgChan, err := e.dbClient.ListPipelineConfigsOfActionExecutor()
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "RELOAD PIPENGINE CONFIG: LIST", err.Error())
	}

	if err := actionexecutor.GetManager().Initialize(cfgChan); err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "RELOAD PIPENGINE CONFIG: RELOAD", err.Error())
	}
	return httpserver.OkResp(nil)
}

func (e *Endpoints) throttlerSnapshot(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	snapshot := e.reconciler.Throttler.Export()
	w.Write(snapshot)
	return nil
}
