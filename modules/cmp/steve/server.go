// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package steve

import (
	"context"
	"errors"
	"net/http"
	"strings"

	apiserver "github.com/rancher/apiserver/pkg/server"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/dynamiclistener/server"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/aggregation"
	"github.com/rancher/steve/pkg/auth"
	"github.com/rancher/steve/pkg/client"
	"github.com/rancher/steve/pkg/clustercache"
	schemacontroller "github.com/rancher/steve/pkg/controllers/schema"
	"github.com/rancher/steve/pkg/resources/common"
	"github.com/rancher/steve/pkg/resources/schemas"
	"github.com/rancher/steve/pkg/schema"
	steveserver "github.com/rancher/steve/pkg/server"
	"github.com/rancher/steve/pkg/server/router"
	"k8s.io/client-go/rest"
)

var ErrConfigRequired = errors.New("rest config is required")

type Server struct {
	http.Handler

	ClientFactory   *client.Factory
	ClusterCache    clustercache.ClusterCache
	SchemaFactory   schema.Factory
	RESTConfig      *rest.Config
	BaseSchemas     *types.APISchemas
	AccessSetLookup accesscontrol.AccessSetLookup
	APIServer       *apiserver.Server
	ClusterRegistry string
	URLPrefix       string

	authMiddleware      auth.Middleware
	controllers         *steveserver.Controllers
	needControllerStart bool
	next                http.Handler
	router              router.RouterFunc

	aggregationSecretNamespace string
	aggregationSecretName      string
}

type Options struct {
	// Controllers If the controllers are passed in the caller must also start the controllers
	Controllers                *steveserver.Controllers
	ClientFactory              *client.Factory
	AccessSetLookup            accesscontrol.AccessSetLookup
	AuthMiddleware             auth.Middleware
	Next                       http.Handler
	Router                     router.RouterFunc
	AggregationSecretNamespace string
	AggregationSecretName      string
	ClusterRegistry            string
	URLPrefix                  string
}

// New create a steve server
func New(ctx context.Context, restConfig *rest.Config, opts *Options) (*Server, error) {
	if opts == nil {
		opts = &Options{}
	}

	host := restConfig.Host
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	restConfig.Host = host
	restConfig.APIPath = "/"

	server := &Server{
		RESTConfig:                 restConfig,
		ClientFactory:              opts.ClientFactory,
		AccessSetLookup:            opts.AccessSetLookup,
		authMiddleware:             opts.AuthMiddleware,
		controllers:                opts.Controllers,
		next:                       opts.Next,
		router:                     opts.Router,
		aggregationSecretNamespace: opts.AggregationSecretNamespace,
		aggregationSecretName:      opts.AggregationSecretName,
		ClusterRegistry:            opts.ClusterRegistry,
		URLPrefix:                  opts.URLPrefix,
	}

	if err := setup(ctx, server); err != nil {
		return nil, err
	}

	return server, server.start(ctx)
}

func setDefaults(server *Server) error {
	if server.RESTConfig == nil {
		return ErrConfigRequired
	}

	if server.controllers == nil {
		var err error
		server.controllers, err = steveserver.NewController(server.RESTConfig, nil)
		server.needControllerStart = true
		if err != nil {
			return err
		}
	}

	if server.next == nil {
		server.next = http.NotFoundHandler()
	}

	if server.BaseSchemas == nil {
		server.BaseSchemas = types.EmptyAPISchemas()
	}

	return nil
}

func setup(ctx context.Context, server *Server) error {
	err := setDefaults(server)
	if err != nil {
		return err
	}

	cf := server.ClientFactory
	if cf == nil {
		cf, err = client.NewFactory(server.RESTConfig, server.authMiddleware != nil)
		if err != nil {
			return err
		}
		server.ClientFactory = cf
	}

	asl := server.AccessSetLookup
	if asl == nil {
		asl = accesscontrol.NewAccessStore(ctx, true, server.controllers.RBAC)
	}

	sf := schema.NewCollection(ctx, server.BaseSchemas, asl)

	DefaultSchemas(server.BaseSchemas)

	k8sInterface, err := cf.AdminK8sInterface()
	if err != nil {
		return err
	}

	for _, template := range DefaultSchemaTemplates(ctx, cf, server.controllers.K8s.Discovery(), asl, k8sInterface) {
		sf.AddTemplate(template)
	}

	cols, err := common.NewDynamicColumns(server.RESTConfig)
	if err != nil {
		return err
	}

	schemas.SetupWatcher(ctx, server.BaseSchemas, asl, sf)

	schemacontroller.Register(ctx,
		cols,
		server.controllers.K8s.Discovery(),
		server.controllers.CRD.CustomResourceDefinition(),
		server.controllers.API.APIService(),
		server.controllers.K8s.AuthorizationV1().SelfSubjectAccessReviews(),
		nil,
		sf)

	apiServer, handler, err := NewHandler(server.RESTConfig, sf, server.authMiddleware, server.next, server.router, server.URLPrefix)
	if err != nil {
		return err
	}

	server.APIServer = apiServer
	server.Handler = handler
	server.SchemaFactory = sf
	return nil
}

func (c *Server) start(ctx context.Context) error {
	if c.needControllerStart {
		if err := c.controllers.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Server) StartAggregation(ctx context.Context) {
	aggregation.Watch(ctx, c.controllers.Core.Secret(), c.aggregationSecretNamespace,
		c.aggregationSecretName, c)
}

func (c *Server) ListenAndServe(ctx context.Context, httpsPort, httpPort int, opts *server.ListenOpts) error {
	if opts == nil {
		opts = &server.ListenOpts{}
	}
	if opts.Storage == nil && opts.Secrets == nil {
		opts.Secrets = c.controllers.Core.Secret()
	}

	c.StartAggregation(ctx)

	if err := server.ListenAndServe(ctx, httpsPort, httpPort, c, opts); err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}
