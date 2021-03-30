package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/pkg/httpserver"
)

// Info 返回 orchestrator 描述
func (e *Endpoints) Info(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: "dice-orchestrator deploys runtime, and provide http request for runtime-related resources.",
	}, nil
}
