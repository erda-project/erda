package trace

import "github.com/erda-project/erda/modules/openapi/api/apis"

var GET_SPANS = apis.ApiSpec{
	Path:        "/api/msp/apm/traces/<traceID>/spans",
	BackendPath: "/api/msp/apm/traces/<traceID>/spans",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "Query spans by traceID.",
}
