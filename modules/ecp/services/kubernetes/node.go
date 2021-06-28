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

// ListNode List node under the specified cluster.
func (k *Kubernetes) ListNode(clusterName string) (*corev1.NodeList, error) {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	nodes, err := client.ClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list node error, %v", clusterName, err)
	}

	return nodes, nil
}

// GetNode Get node under the specified cluster.
func (k *Kubernetes) GetNode(clusterName, nodeName string) (*corev1.Node, error) {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	np, err := client.ClientSet.CoreV1().Nodes().Get(context.Background(), nodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get node %s error, %v", clusterName, nodeName, err)
	}

	return np, nil
}

// CreateNode Create node on specified cluster.
func (k *Kubernetes) CreateNode(clusterName string, node *corev1.Node) (*corev1.Node, error) {
	if node == nil {
		return nil, fmt.Errorf("create action must give a non-nil node entity")
	}

	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	res, err := client.ClientSet.CoreV1().Nodes().Create(context.Background(), node, v1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s create node %s error, %v", clusterName, node.Name, err)
	}

	return res, nil
}

// DeleteNode Delete node on specified cluster.
func (k *Kubernetes) DeleteNode(clusterName string, nodeName string) error {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = client.ClientSet.CoreV1().Nodes().Delete(context.Background(), nodeName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete node error, %v", clusterName, err)
	}

	return nil
}

// UpdateNode Update node on specified cluster.
func (k *Kubernetes) UpdateNode(clusterName string, node *corev1.Node) error {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = client.ClientSet.CoreV1().Nodes().Update(context.Background(), node, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update node error, %v", clusterName, err)
	}

	return nil
}
