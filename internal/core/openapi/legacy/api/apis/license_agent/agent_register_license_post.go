package license_agent

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
	"net/http"
)

var AGENT_REGISTER_LICENSE_POST = apis.ApiSpec{
	Path:         "/api/licenses/actions/register",
	BackendPath:  "/api/licenses/actions/register",
	Host:         "license-agent.marathon.l4lb.thisdcos.directory:8080",
	Scheme:       "http",
	Method:       http.MethodPost,
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.RegisterLicenseRequest{},
	ResponseType: apistructs.RegisterLicenseResponse{},
}
