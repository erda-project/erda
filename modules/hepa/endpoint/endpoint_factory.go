// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package endpoint

import (
	"github.com/erda-project/erda/modules/hepa/k8s"
)

type Route struct {
	Host string
	Path string
}

type EndpointMaterial struct {
	K8SNamespace          string
	ServiceName           string
	ServicePort           int
	Routes                []Route
	ServiceGroupNamespace string
	ServiceGroupName      string
	ProjectNamespace      string
	IngressName           string
	K8SRouteOptions       k8s.RouteOptions
}

type EndpointFactory interface {
	TouchRoute(EndpointMaterial) error
	ClearRoute(EndpointMaterial) error
	TouchComponentRoute(EndpointMaterial) error
	ClearComponentRoute(EndpointMaterial) error
}

func TouchEndpoint(factory EndpointFactory, material EndpointMaterial) error {
	if len(material.Routes) == 0 {
		return factory.ClearRoute(material)
	}
	return factory.TouchRoute(material)
}

func TouchComponentEndpoint(factory EndpointFactory, material EndpointMaterial) error {
	if len(material.Routes) == 0 {
		return factory.ClearComponentRoute(material)
	}
	return factory.TouchComponentRoute(material)
}
