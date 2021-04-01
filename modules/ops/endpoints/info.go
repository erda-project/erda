package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/pkg/httpserver"
)

// Info demo
func (e *Endpoints) Info(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	return httpserver.OkResp("ok")
}
