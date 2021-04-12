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

// ListDeployment List deployment under the specified cluster and namespace.
func (k *Kubernetes) ListDeployment(clusterName, namespace string) (*appsv1.DeploymentList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cmList, err := clientSet.K8sClient.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list deployment(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return cmList, nil
}

// GetDeployment Get deployment under the specified cluster and namespace.
func (k *Kubernetes) GetDeployment(clusterName, namespace, deploymentName string) (*appsv1.Deployment, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := clientSet.K8sClient.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get deployment(namespace: %s) %s error, %v", clusterName, namespace, deploymentName, err)
	}

	return cm, nil
}

// CreateDeployment Crate deployment on specified cluster and namespace.
func (k *Kubernetes) CreateDeployment(clusterName, namespace string, cm *appsv1.Deployment) error {
	if cm == nil {
		return fmt.Errorf("deployment entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.AppsV1().Deployments(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create deployment(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}

// DeleteDeployment Delete deployment on specified cluster and namespace.
func (k *Kubernetes) DeleteDeployment(clusterName, namespace, deploymentName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete deployment(namespace: %s) %s error, %v", clusterName, namespace, deploymentName, err)
	}

	return nil
}

// UpdateDeployment Update deployment on specified cluster and namespace.
func (k *Kubernetes) UpdateDeployment(clusterName, namespace string, cm *appsv1.Deployment) error {
	if cm == nil {
		return fmt.Errorf("deployment entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.AppsV1().Deployments(namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update deployment(namespace: %s) %s error, %v", clusterName, namespace, cm.Name, err)
	}

	return nil
}
