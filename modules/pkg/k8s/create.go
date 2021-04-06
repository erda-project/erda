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

package k8s

import (
	"bytes"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"github.com/erda-project/erda/pkg/httpclient"
)

// CreateDaemonSet 通过 API Server 地址创建 DaemonSet
func CreateDaemonSet(as string, ds appsv1.DaemonSet) error {
	var b bytes.Buffer
	resp, err := httpclient.New().
		Post(as).
		Path("/apis/apps/v1/namespaces/" + ds.Namespace + "/daemonsets").
		JSONBody(ds).Do().Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create daemonset, namespace: %s, name: %s, (%s)\n\t%s",
			ds.Namespace, ds.Name, err.Error(), b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create daemonset, namespace: %s, name: %s, (status code is %d)\n\t%s",
			ds.Namespace, ds.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// CreateDeployment 通过 API Server 地址创建 Deployment
func CreateDeployment(as string, deploy appsv1.Deployment) error {
	var b bytes.Buffer
	resp, err := httpclient.New().
		Post(as).
		Path("/apis/apps/v1/namespaces/" + deploy.Namespace + "/deployments").
		JSONBody(deploy).Do().Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create deployment, namespace: %s, name: %s, (%s)\n\t%s",
			deploy.Namespace, deploy.Name, err.Error(), b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create deployment, namespace: %s, name: %s, (status code is %d)\n\t%s",
			deploy.Namespace, deploy.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// CreateService 通过 API Server 地址创建 Service
func CreateService(as string, svc corev1.Service) error {
	var b bytes.Buffer
	resp, err := httpclient.New().
		Post(as).
		Path("/api/v1/namespaces/" + svc.Namespace + "/services").
		JSONBody(svc).Do().Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create service, namespace: %s, name: %s, (%s)\n\t%s",
			svc.Namespace, svc.Name, err.Error(), b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create service, namespace: %s, name: %s, (status code is %d)\n\t%s",
			svc.Namespace, svc.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// CreateIngress 通过 API Server 地址创建 Ingress
func CreateIngress(as string, ingress extensionsv1beta1.Ingress) error {
	var b bytes.Buffer
	resp, err := httpclient.New().
		Post(as).
		Path("/apis/extensions/v1beta1/namespaces/" + ingress.Namespace + "/ingresses").
		JSONBody(ingress).Do().Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create ingress, namespace: %s, name: %s, (%s)\n\t%s",
			ingress.Namespace, ingress.Name, err.Error(), b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create ingress, namespace: %s, name: %s, (status code is %d)\n\t%s",
			ingress.Namespace, ingress.Name, resp.StatusCode(), b.String())
	}
	return nil
}
