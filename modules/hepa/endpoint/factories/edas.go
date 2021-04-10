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

type EdasFactory struct {
	k8sAdapter k8s.K8SAdapter
}

func NewEdasFactory(masterAddr string) (endpoint.EndpointFactory, error) {
	adapter, err := k8s.NewAdapter(masterAddr)
	if err != nil {
		return nil, err
	}
	return EdasFactory{
		k8sAdapter: adapter,
	}, nil
}

func (impl EdasFactory) serviceName(serviceGroupNamespace, serviceGroupName, serviceName string) string {
	return serviceGroupNamespace + "-" + serviceGroupName + "-" + serviceName
}

func (impl EdasFactory) TouchRoute(material endpoint.EndpointMaterial) error {
	var k8sRoutes []k8s.IngressRoute
	for _, route := range material.Routes {
		k8sRoutes = append(k8sRoutes, k8s.IngressRoute{
			Domain: route.Host,
			Path:   route.Path,
		})
	}
	serviceName := impl.serviceName(material.ServiceGroupNamespace, material.ServiceGroupName, material.ServiceName)
	backend := k8s.IngressBackend{
		ServiceName: serviceName,
		ServicePort: material.ServicePort,
	}
	_, err := impl.k8sAdapter.CreateOrUpdateIngress("default", serviceName,
		k8sRoutes, backend, material.K8SRouteOptions)
	if err != nil {
		return err
	}
	return nil
}

func (impl EdasFactory) TouchComponentRoute(material endpoint.EndpointMaterial) error {
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

func (impl EdasFactory) ClearComponentRoute(material endpoint.EndpointMaterial) error {
	namespace := "default"
	if material.K8SNamespace != "" {
		namespace = material.K8SNamespace
	}
	return impl.k8sAdapter.DeleteIngress(namespace, material.IngressName)
}
func (impl EdasFactory) ClearRoute(material endpoint.EndpointMaterial) error {
	return impl.k8sAdapter.DeleteIngress("default", impl.serviceName(material.ServiceGroupNamespace, material.ServiceGroupName, material.ServiceName))
}
