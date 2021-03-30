package endpoints

import (
	"context"
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// MigrationLog migration log接口
func (e *Endpoints) MigrationLog(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	//鉴权
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateAddon.NotLogin().ToResp(), nil
	}
	// 获取日志
	var logReq apistructs.DashboardSpotLogRequest
	if err := queryStringDecoder.Decode(&logReq, r.URL.Query()); err != nil {
		return apierrors.ErrMigrationLog.InvalidParameter(err).ToResp(), nil
	}
	logReq.ID = "migration-task-" + vars["migrationId"]
	logReq.Source = apistructs.DashboardSpotLogSourceJob
	// 查询日志信息
	log, err := e.bdl.GetLog(logReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if len(log.Lines) == 0 {
		log.Lines = []apistructs.DashboardSpotLogLine{}
	}
	return httpserver.OkResp(log)
}

// 清理migration namespace
func (e *Endpoints) CleanUnusedMigrationNs() (bool, error) {
	return e.migration.CleanUnusedMigrationNs()
}

var queryStringDecoder *schema.Decoder

func init() {
	queryStringDecoder = schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
}
