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

// ListConfigMap List configMap under the specified cluster and namespace.
func (k *Kubernetes) ListConfigMap(clusterName, namespace string) (*corev1.ConfigMapList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cmList, err := clientSet.K8sClient.CoreV1().ConfigMaps(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list configmap(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return cmList, nil
}

// GetConfigMap Get configMap under the specified cluster and namespace.
func (k *Kubernetes) GetConfigMap(clusterName, namespace, cmName string) (*corev1.ConfigMap, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := clientSet.K8sClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get configmap(namespace: %s) %s error, %v", clusterName, namespace, cmName, err)
	}

	return cm, nil
}

// CreateConfigMap Crate configMap on specified cluster and namespace.
func (k *Kubernetes) CreateConfigMap(clusterName, namespace string, cm *corev1.ConfigMap) error {
	if cm == nil {
		return fmt.Errorf("configmap entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create configmap(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}

// DeleteConfigMap Delete configMap on specified cluster and namespace.
func (k *Kubernetes) DeleteConfigMap(clusterName, namespace, cmName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), cmName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete configmap(namespace: %s) %s error, %v", clusterName, namespace, cmName, err)
	}

	return nil
}

// UpdateConfigMap Update configMap on specified cluster and namespace.
func (k *Kubernetes) UpdateConfigMap(clusterName, namespace string, cm *corev1.ConfigMap) error {
	if cm == nil {
		return fmt.Errorf("configmap entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update configmap(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}
