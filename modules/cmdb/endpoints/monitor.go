package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/modules/cmdb/services/monitor"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) RunIssueHistory(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	go monitor.RunIssueHistoryData(e.DBClient(), e.uc, e.bdl)
	return httpserver.OkResp(nil)
}

func (e *Endpoints) RunIssueAddOrRepairHistory(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	go monitor.RunHistoryData(e.DBClient(), e.bdl)
	return httpserver.OkResp(nil)
}
