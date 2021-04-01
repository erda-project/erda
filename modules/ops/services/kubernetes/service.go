package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetService 获取指定集群下 Service
func (k *Kubernetes) GetService(clusterName, namespace, svcName string) (*corev1.Service, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	cm, err := clientSet.K8sClient.CoreV1().Services(namespace).Get(context.TODO(), svcName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get service(namespace: %s) %s error, %v", clusterName, namespace, svcName, err)
	}

	return cm, nil
}

// CreateService 在指定集群下创建 Service
func (k *Kubernetes) CreateService(clusterName, namespace string, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("service entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Services(namespace).Create(context.TODO(), svc, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create service(namespace: %s) %s error, %v", clusterName, namespace, svc.Name, err)
	}

	return nil
}

// DeleteService 在指定集群下删除 Service
func (k *Kubernetes) DeleteService(clusterName, namespace, svcName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().Services(namespace).Delete(context.TODO(), svcName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete service(namespace: %s) %s error, %v", clusterName, namespace, svcName, err)
	}

	return nil
}

// UpdateService 在指定集群下更新 Service
func (k *Kubernetes) UpdateService(clusterName, namespace string, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Services(namespace).Update(context.TODO(), svc, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update service(namespace: %s) %s error, %v", clusterName, namespace, svc.Name, err)
	}

	return nil
}
