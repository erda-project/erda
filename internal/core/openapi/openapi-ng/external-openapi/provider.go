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

package external_openapi

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/grpcserver"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes"
)

var (
	_ pb.DynamicOpenapiRegisterServer = (*provider)(nil)
)

var (
	name = "erda.openapi-ng.external-openapi"
	spec = servicehub.Spec{
		Summary:     "external apis expose in erda openapi",
		Description: "external apis expose in erda openapi",
		ConfigFunc:  func() any { return new(struct{}) },
		Creator:     func() servicehub.Provider { return new(provider) },
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Log     logs.Logger
	Openapi routes.Register      `autowired:"openapi-dynamic-register.client"`
	GRPC    grpcserver.Interface `autowired:"grpc-server"`
}

func (p *provider) Init(_ servicehub.Context) error {
	pb.RegisterDynamicOpenapiRegisterServer(p.GRPC, p)
	return nil
}

func (p *provider) Register(ctx context.Context, api *pb.API) (*common.VoidResponse, error) {
	if err := adjust(api); err != nil {
		return nil, err
	}
	proxy := &routes.APIProxy{
		Method:      api.GetMethod(),
		Path:        api.GetPath(),
		ServiceURL:  api.GetUpstream(),
		BackendPath: path.Join(path.Clean("/"+api.GetModule()), path.Clean("/"+api.GetBackendPath())),
		Auth:        api.GetAuth(),
	}
	p.Log.Infof("register external API to openapi, %s %s -> %s%s\n", proxy.Method, proxy.Path, proxy.ServiceURL, proxy.BackendPath)
	if err := p.Openapi.Register(proxy); err != nil {
		return nil, errors.Wrapf(err, "failed to register the api %+v to openapi", api)
	}
	return new(common.VoidResponse), nil
}

func (p *provider) Deregister(ctx context.Context, api *pb.API) (*common.VoidResponse, error) {
	return nil, errors.New("not implement")
}

func adjust(api *pb.API) error {
	u, err := url.Parse(api.GetUpstream())
	if err != nil {
		return errors.Wrapf(err, "api host %s is invalid", api.GetUpstream())
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	api.Upstream = u.String()
	if api.GetMethod() == "" {
		api.Method = "*"
	}
	api.Method = strings.ToUpper(api.GetMethod())
	if api.GetPath() == "" {
		api.Path = "/"
	}
	if api.GetBackendPath() == "" {
		api.BackendPath = api.GetPath()
	}
	return nil
}
