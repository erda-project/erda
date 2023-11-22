package license_agent

import (
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
	"net/http"
)

var AGENT_LICENSE_GET = apis.ApiSpec{
	Path:        "/api/licenses",
	BackendPath: "/api/licenses",
	Host:        "license-agent.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      http.MethodGet,
	IsOpenAPI:   true,
	CheckLogin:  true,
	CheckToken:  true,
}
