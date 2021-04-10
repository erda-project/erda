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

package factories

import (
	"github.com/erda-project/erda/modules/hepa/endpoint"
	"github.com/erda-project/erda/modules/hepa/k8s"
)

type K8SFactory struct {
	k8sAdapter k8s.K8SAdapter
}

func NewK8SFactory(masterAddr string) (endpoint.EndpointFactory, error) {
	adapter, err := k8s.NewAdapter(masterAddr)
	if err != nil {
		return nil, err
	}
	return K8SFactory{
		k8sAdapter: adapter,
	}, nil
}

func (impl K8SFactory) k8sNamespace(material endpoint.EndpointMaterial) string {
	if material.ProjectNamespace != "" {
		return material.ProjectNamespace
	}
	return material.ServiceGroupNamespace + "--" + material.ServiceGroupName
}

func (impl K8SFactory) k8sServiceName(material endpoint.EndpointMaterial) string {
	if material.ProjectNamespace == "" {
		return material.ServiceName
	}
	return material.ServiceName + "-" + material.ServiceGroupName
}

func (impl K8SFactory) TouchRoute(material endpoint.EndpointMaterial) error {
	var k8sRoutes []k8s.IngressRoute
	for _, route := range material.Routes {
		k8sRoutes = append(k8sRoutes, k8s.IngressRoute{
			Domain: route.Host,
			Path:   route.Path,
		})
	}
	serviceName := impl.k8sServiceName(material)
	backend := k8s.IngressBackend{
		ServiceName: serviceName,
		ServicePort: material.ServicePort,
	}
	_, err := impl.k8sAdapter.CreateOrUpdateIngress(impl.k8sNamespace(material),
		serviceName, k8sRoutes, backend, material.K8SRouteOptions)
	if err != nil {
		return err
	}
	return nil
}

func (impl K8SFactory) TouchComponentRoute(material endpoint.EndpointMaterial) error {
	var k8sRoutes []k8s.IngressRoute
	for _, route := range material.Routes {
		k8sRoutes = append(k8sRoutes, k8s.IngressRoute{
			Domain: route.Host,
			Path:   route.Path,
		})
	}
	backend := k8s.IngressBackend{
		ServiceName: material.ServiceName,
		ServicePort: material.ServicePort,
	}
	namespace := "default"
	if material.K8SNamespace != "" {
		namespace = material.K8SNamespace
	}
	_, err := impl.k8sAdapter.CreateOrUpdateIngress(namespace, material.IngressName, k8sRoutes, backend, material.K8SRouteOptions)
	if err != nil {
		return err
	}
	return nil
}

func (impl K8SFactory) ClearComponentRoute(material endpoint.EndpointMaterial) error {
	namespace := "default"
	if material.K8SNamespace != "" {
		namespace = material.K8SNamespace
	}
	return impl.k8sAdapter.DeleteIngress(namespace, material.IngressName)
}
func (impl K8SFactory) ClearRoute(material endpoint.EndpointMaterial) error {
	return impl.k8sAdapter.DeleteIngress(impl.k8sNamespace(material),
		impl.k8sServiceName(material))
}
