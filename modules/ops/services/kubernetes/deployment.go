package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployment 获取指定集群下指定 namespace 所有 deployment
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

// GetDeployment 获取指定集群下 deployment
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

// CreateDeployment 在指定集群下创建 deployment
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

// DeleteDeployment 在指定集群下删除 deployment
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

// UpdateDeployment 在指定集群下更新 deployment
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
