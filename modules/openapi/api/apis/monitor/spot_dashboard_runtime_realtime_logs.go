package monitor

import "github.com/erda-project/erda/modules/openapi/api/apis"

var SPOT_RUNTIME_REALTIME_LOGS = apis.ApiSpec{
	Path:        "/api/runtime/realtime/logs",
	BackendPath: "/api/runtime/realtime/logs",
	Host:        "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 查询Runtime实时日志内容",
}
