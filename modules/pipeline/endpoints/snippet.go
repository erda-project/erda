package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) querySnippetDetails(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var req apistructs.SnippetQueryDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	data, err := e.snippetSvc.QueryDetails(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}
