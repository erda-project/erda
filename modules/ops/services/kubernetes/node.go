package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNode 获取 某个 指定 clusterName 集群下的所有 Node
func (k *Kubernetes) ListNode(clusterName string) (*corev1.NodeList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	nodes, err := clientSet.K8sClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list node error, %v", clusterName, err)
	}

	return nodes, nil
}

// GetNode 获取指定 clusterName 下的 Node
func (k *Kubernetes) GetNode(clusterName, nodeName string) (*corev1.Node, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	np, err := clientSet.K8sClient.CoreV1().Nodes().Get(context.Background(), nodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get node %s error, %v", clusterName, nodeName, err)
	}

	return np, nil
}

// CreateNode 指定 clusterName 集群下创建 Node
func (k *Kubernetes) CreateNode(clusterName string, node *corev1.Node) (*corev1.Node, error) {
	if node == nil {
		return nil, fmt.Errorf("create action must give a non-nil node entity")
	}

	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	res, err := clientSet.K8sClient.CoreV1().Nodes().Create(context.TODO(), node, v1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s create node %s error, %v", clusterName, node.Name, err)
	}

	return res, nil
}

// DeleteNode 指定 clusterName 集群下创建 Node
func (k *Kubernetes) DeleteNode(clusterName string, nodeName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().Nodes().Delete(context.TODO(), nodeName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete node error, %v", clusterName, err)
	}

	return nil
}

// UpdateNode 指定 clusterName 集群下更新 Node
func (k *Kubernetes) UpdateNode(clusterName string, node *corev1.Node) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Nodes().Update(context.TODO(), node, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update node error, %v", clusterName, err)
	}

	return nil
}
