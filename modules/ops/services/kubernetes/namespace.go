package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace 指定集群下创建 namespace
func (k *Kubernetes) CreateNamespace(clusterName string, namespace *corev1.Namespace) error {
	if namespace == nil {
		return fmt.Errorf("namespace entity can't be nil")
	}

	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Namespaces().Create(context.TODO(), namespace, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create namespace %s error, %v", clusterName, namespace.Name, err)
	}
	return nil
}

// DeleteNamespace 指定集群下删除 namespace
func (k *Kubernetes) DeleteNamespace(clusterName string, namespace string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete namespace %s error, %v", clusterName, namespace, err)
	}

	return nil
}
