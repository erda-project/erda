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

// ListSecret 获取指定集群下指定 namespace 所有 Secret
func (k *Kubernetes) ListSecret(clusterName, namespace string) (*corev1.SecretList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cmList, err := clientSet.K8sClient.CoreV1().Secrets(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list secret(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return cmList, nil
}

// GetSecret 获取指定集群下 Secret
func (k *Kubernetes) GetSecret(clusterName, namespace, secName string) (*corev1.Secret, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := clientSet.K8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), secName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get secret(namespace: %s) %s error, %v", clusterName, namespace, secName, err)
	}

	return cm, nil
}

// CreateSecret 在指定集群下创建 Secret
func (k *Kubernetes) CreateSecret(clusterName, namespace string, cm *corev1.Secret) error {
	if cm == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Secrets(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create secret(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}

// DeleteSecret 在指定集群下删除 Secret
func (k *Kubernetes) DeleteSecret(clusterName, namespace, secName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().Secrets(namespace).Delete(context.TODO(), secName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete Secret(namespace: %s) %s error, %v", clusterName, namespace, secName, err)
	}

	return nil
}

// UpdateSecret 在指定集群下更新 Secret
func (k *Kubernetes) UpdateSecret(clusterName, namespace string, cm *corev1.Secret) error {
	if cm == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Secrets(namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update secret(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}
