package steve

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/apiserver/pkg/server"
	apiserver "github.com/rancher/apiserver/pkg/server"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/steve/pkg/auth"
	k8sproxy "github.com/rancher/steve/pkg/proxy"
	"github.com/rancher/steve/pkg/schema"
	"github.com/rancher/steve/pkg/server/router"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/strutil"
)

// GetURLPrefix get steve API prefix with cluster name
func GetURLPrefix(clusterName string) string {
	return strutil.Concat("/k8s/clusters/", clusterName)
}

// NewHandler return an rancher api server and a steve server handler
func NewHandler(cfg *rest.Config, sf schema.Factory, authMiddleware auth.Middleware, next http.Handler,
	routerFunc router.RouterFunc, prefix string) (*apiserver.Server, http.Handler, error) {
	var (
		proxy http.Handler
		err   error
	)

	a := &apiServer{
		sf:     sf,
		server: server.DefaultAPIServer(),
	}
	a.server.AccessControl = accesscontrol.NewAccessControl()

	if authMiddleware == nil {
		proxy, err = k8sproxy.Handler(prefix, cfg)
		if err != nil {
			return a.server, nil, err
		}
		authMiddleware = auth.ToMiddleware(auth.AuthenticatorFunc(auth.AlwaysAdmin))
	} else {
		proxy = k8sproxy.ImpersonatingHandler(prefix, cfg)
	}

	w := authMiddleware
	handlers := router.Handlers{
		Next:        next,
		K8sResource: w(a.apiHandler(k8sAPI, prefix)),
		K8sProxy:    w(proxy),
		APIRoot:     w(a.apiHandler(apiRoot, prefix)),
	}
	if routerFunc == nil {
		return a.server, router.Routes(handlers), nil
	}
	return a.server, routerFunc(handlers), nil
}

type apiServer struct {
	sf     schema.Factory
	server *server.Server
}

func (a *apiServer) common(rw http.ResponseWriter, req *http.Request, prefix string) (*types.APIRequest, bool) {
	user, ok := request.UserFrom(req.Context())
	if !ok {
		return nil, false
	}

	schemas, err := a.sf.Schemas(user)
	if err != nil {
		logrus.Errorf("HTTP request failed: %v", err)
		rw.Write([]byte(err.Error()))
		rw.WriteHeader(http.StatusInternalServerError)
	}

	prefixedUrlBuilder, err := NewPrefixed(req, schemas, prefix)
	if err != nil {
		rw.Write([]byte(err.Error()))
		rw.WriteHeader(http.StatusInternalServerError)
		return nil, false
	}

	return &types.APIRequest{
		Schemas:    schemas,
		Request:    req,
		Response:   rw,
		URLBuilder: prefixedUrlBuilder,
	}, true
}

type APIFunc func(schema.Factory, *types.APIRequest)

func (a *apiServer) apiHandler(apiFunc APIFunc, prefix string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		apiOp, ok := a.common(rw, req, prefix)
		if ok {
			if apiFunc != nil {
				apiFunc(a.sf, apiOp)
			}
			a.server.Handle(apiOp)
		}
	})
}

func k8sAPI(sf schema.Factory, apiOp *types.APIRequest) {
	vars := mux.Vars(apiOp.Request)
	group := vars["group"]
	if group == "core" {
		group = ""
	}

	apiOp.Name = vars["name"]
	apiOp.Type = vars["type"]

	nOrN := vars["nameorns"]
	if nOrN != "" {
		schema := apiOp.Schemas.LookupSchema(apiOp.Type)
		if attributes.Namespaced(schema) {
			vars["namespace"] = nOrN
		} else {
			vars["name"] = nOrN
		}
	}

	if namespace := vars["namespace"]; namespace != "" {
		apiOp.Namespace = namespace
	}
}

func apiRoot(sf schema.Factory, apiOp *types.APIRequest) {
	apiOp.Type = "apiRoot"
}
