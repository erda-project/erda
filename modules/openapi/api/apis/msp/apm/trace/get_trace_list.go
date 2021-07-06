package trace

import "github.com/erda-project/erda/modules/openapi/api/apis"

var GET_TRACE_LIST = apis.ApiSpec{
	Path:        "/api/msp/apm/traces",
	BackendPath: "/api/msp/apm/traces",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "Query apm traces.",
}
