package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/version"
)

func (e *Endpoints) version(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return httpserver.OkResp(version.String())
}
