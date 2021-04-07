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

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListStatefulSet 获取指定集群下指定 namespace 所有 statefulSet
func (k *Kubernetes) ListStatefulSet(clusterName, namespace string) (*appsv1.StatefulSetList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cmList, err := clientSet.K8sClient.AppsV1().StatefulSets(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list statefulSet(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return cmList, nil
}

// GetStatefulSet 获取指定集群下 statefulSet
func (k *Kubernetes) GetStatefulSet(clusterName, namespace, statefulSetName string) (*appsv1.StatefulSet, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := clientSet.K8sClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get statefulSet(namespace: %s) %s error, %v", clusterName, namespace, statefulSetName, err)
	}

	return cm, nil
}

// CreateStatefulSet 在指定集群下创建 statefulSet
func (k *Kubernetes) CreateStatefulSet(clusterName, namespace string, cm *appsv1.StatefulSet) error {
	if cm == nil {
		return fmt.Errorf("statefulSet entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.AppsV1().StatefulSets(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create statefulSet(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}

// DeleteStatefulSet 在指定集群下删除 statefulSet
func (k *Kubernetes) DeleteStatefulSet(clusterName, namespace, statefulSetName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), statefulSetName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete statefulSet(namespace: %s) %s error, %v", clusterName, namespace, statefulSetName, err)
	}

	return nil
}

// UpdateStatefulSet 在指定集群下更新 statefulSet
func (k *Kubernetes) UpdateStatefulSet(clusterName, namespace string, cm *appsv1.StatefulSet) error {
	if cm == nil {
		return fmt.Errorf("statefulSet entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.AppsV1().StatefulSets(namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update statefulSet(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}
