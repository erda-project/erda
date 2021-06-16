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

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetService List service resource under the specified cluster and namespace.
func (k *Kubernetes) GetService(clusterName, namespace, svcName string) (*corev1.Service, error) {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := client.ClientSet.CoreV1().Services(namespace).Get(context.Background(), svcName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get service(namespace: %s) %s error, %v", clusterName, namespace, svcName, err)
	}

	return cm, nil
}

// CreateService Crate service on specified cluster and namespace.
func (k *Kubernetes) CreateService(clusterName, namespace string, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("service entity can't be nil")
	}
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = client.ClientSet.CoreV1().Services(namespace).Create(context.Background(), svc, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create service(namespace: %s) %s error, %v", clusterName, namespace, svc.Name, err)
	}

	return nil
}

// DeleteService Delete service on specified cluster and namespace.
func (k *Kubernetes) DeleteService(clusterName, namespace, svcName string) error {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = client.ClientSet.CoreV1().Services(namespace).Delete(context.Background(), svcName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete service(namespace: %s) %s error, %v", clusterName, namespace, svcName, err)
	}

	return nil
}

// UpdateService Update service on specified cluster and namespace.
func (k *Kubernetes) UpdateService(clusterName, namespace string, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = client.ClientSet.CoreV1().Services(namespace).Update(context.Background(), svc, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update service(namespace: %s) %s error, %v", clusterName, namespace, svc.Name, err)
	}

	return nil
}
