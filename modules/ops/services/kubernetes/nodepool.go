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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/clientgo/apis/openyurt/v1alpha1"
)

// ListNodePool 获取 某个 指定 clusterName 集群下的所有 NodePool
func (k *Kubernetes) ListNodePool(clusterName string) (*v1alpha1.NodePoolList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	nodePools, err := clientSet.CustomClient.OpenYurtV1alpha1().NodePools().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list nodepool error, %v", clusterName, err)
	}

	return nodePools, nil
}

// GetNodePool 获取指定 clusterName 下的 NodePool
func (k *Kubernetes) GetNodePool(clusterName, npName string) (*v1alpha1.NodePool, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	np, err := clientSet.CustomClient.OpenYurtV1alpha1().NodePools().Get(context.Background(), npName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get nodepool %s error, %v", clusterName, npName, err)
	}

	return np, nil
}

// CreateNodePool 指定 clusterName 集群下创建 NodePool
func (k *Kubernetes) CreateNodePool(clusterName string, nodePool *v1alpha1.NodePool) (*v1alpha1.NodePool, error) {
	if nodePool == nil {
		return nil, fmt.Errorf("create action must give a non-nil NodePool entity")
	}

	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	res, err := clientSet.CustomClient.OpenYurtV1alpha1().NodePools().Create(context.TODO(), nodePool)
	if err != nil {
		return nil, fmt.Errorf("cluster %s create nodepool %s error, %v", clusterName, nodePool.Name, err)
	}

	return res, nil
}

// DeleteNodePool 指定 clusterName 集群下创建 NodePool
func (k *Kubernetes) DeleteNodePool(clusterName string, nodePoolName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.CustomClient.OpenYurtV1alpha1().NodePools().Delete(context.TODO(),
		nodePoolName, &v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete nodepool error, %v", clusterName, err)
	}

	return nil
}

// UpdateNodePool 指定 clusterName 集群下更新 NodePool
func (k *Kubernetes) UpdateNodePool(clusterName string, nodePool *v1alpha1.NodePool) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.CustomClient.OpenYurtV1alpha1().NodePools().Update(context.TODO(), nodePool)
	if err != nil {
		return fmt.Errorf("cluster %s update nodepool error, %v", clusterName, err)
	}

	return nil
}
