package endpoints

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/pkg/httpserver"
)

// 创建service元数据，由调度器来填写
func (e *Endpoints) serviceCreateOrUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		service types.CmService
		body    []byte
		err     error
	)

	if body, err = ioutil.ReadAll(r.Body); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: "read http request body error."}, err
	}

	if err = json.Unmarshal(body, &service); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: "unmarshal failed."}, err
	}

	service.Cluster = vars["cluster"]
	service.DiceProject = vars["project"]
	service.DiceApplication = vars["application"]
	service.DiceRuntime = vars["runtime"]
	service.DiceService = vars["service"]

	if err = e.db.CreateOrUpdateService(ctx, &service); err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError, Content: "update service to database failed."}, err
	}

	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: &service,
	}, err
}
