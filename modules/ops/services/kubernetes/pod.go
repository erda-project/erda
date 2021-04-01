package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPod 获取指定集群下 Pod
func (k *Kubernetes) ListPod(clusterName, namespace string) (*corev1.PodList, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	po, err := clientSet.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list pod(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return po, nil
}

// GetPod 获取指定集群下 Pod
func (k *Kubernetes) GetPod(clusterName, namespace, podName string) (*corev1.Pod, error) {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	po, err := clientSet.K8sClient.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get pod(namespace: %s) %s error, %v", clusterName, namespace, podName, err)
	}

	return po, nil
}

// CreatePod 在指定集群下创建 Pod
func (k *Kubernetes) CreatePod(clusterName, namespace string, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("pod entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create pod(namespace: %s) %s error, %v", clusterName, namespace, pod.Name, err)
	}

	return nil
}

// DeletePod 在指定集群下删除 Pod
func (k *Kubernetes) DeletePod(clusterName, namespace, podName string) error {
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = clientSet.K8sClient.CoreV1().Pods(namespace).Delete(context.TODO(), podName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete pod(namespace: %s) %s error, %v", clusterName, namespace, podName, err)
	}

	return nil
}

// UpdatePod 在指定集群下更新 Pod
func (k *Kubernetes) UpdatePod(clusterName, namespace string, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	clientSet, err := k.getClientSet(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = clientSet.K8sClient.CoreV1().Pods(namespace).Update(context.TODO(), pod, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update pod(namespace: %s) %s error, %v", clusterName, namespace, pod.Name, err)
	}

	return nil
}
