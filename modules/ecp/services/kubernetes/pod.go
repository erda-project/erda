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

// GetPod List pod under the specified cluster and namespace.
func (k *Kubernetes) ListPod(clusterName, namespace string) (*corev1.PodList, error) {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	po, err := client.ClientSet.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s list pod(namespace: %s) error, %v", clusterName, namespace, err)
	}

	return po, nil
}

// GetPod Get pod under the specified cluster and namespace.
func (k *Kubernetes) GetPod(clusterName, namespace, podName string) (*corev1.Pod, error) {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	po, err := client.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), podName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cluster %s get pod(namespace: %s) %s error, %v", clusterName, namespace, podName, err)
	}

	return po, nil
}

// CreatePod Crate pod on specified cluster and namespace.
func (k *Kubernetes) CreatePod(clusterName, namespace string, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("pod entity can't be nil")
	}
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = client.ClientSet.CoreV1().Pods(namespace).Create(context.Background(), pod, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s create pod(namespace: %s) %s error, %v", clusterName, namespace, pod.Name, err)
	}

	return nil
}

// DeletePod Delete pod on specified cluster and namespace.
func (k *Kubernetes) DeletePod(clusterName, namespace, podName string) error {
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	err = client.ClientSet.CoreV1().Pods(namespace).Delete(context.Background(), podName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s delete pod(namespace: %s) %s error, %v", clusterName, namespace, podName, err)
	}

	return nil
}

// UpdatePod Update pod on specified cluster and namespace.
func (k *Kubernetes) UpdatePod(clusterName, namespace string, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("secret entity can't be nil")
	}
	client, err := k.getK8sClient(clusterName)
	if err != nil {
		return fmt.Errorf("get cluster %s clientset error, %v", clusterName, err)
	}

	_, err = client.ClientSet.CoreV1().Pods(namespace).Update(context.Background(), pod, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cluster %s update pod(namespace: %s) %s error, %v", clusterName, namespace, pod.Name, err)
	}

	return nil
}
