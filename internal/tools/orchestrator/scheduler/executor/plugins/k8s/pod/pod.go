// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package pod manipulates the k8s api of pod object
package pod

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	defaultPodLimit = 10
)

// Pod is the object to encapsulate docker
type Pod struct {
	k8sClient kubernetes.Interface
}

// Option configures a Pod
type Option func(*Pod)

// New news a Pod
func New(options ...Option) *Pod {
	ns := &Pod{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

func WithK8sClient(k8sClient kubernetes.Interface) Option {
	return func(p *Pod) {
		p.k8sClient = k8sClient
	}
}

// Get gets a k8s pod
func (p *Pod) Get(namespace, name string) (*corev1.Pod, error) {
	pod, err := p.k8sClient.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, k8serror.ErrNotFound
		}
		return nil, err
	}

	return pod, nil
}

func (p *Pod) Delete(namespace, name string) error {
	err := p.k8sClient.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (p *Pod) WatchAllNamespace(ctx context.Context, addFunc, modifyFunc, delFunc func(*corev1.Pod)) error {
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceSystem))
	pods, err := p.k8sClient.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		Limit: defaultPodLimit,
	})
	if err != nil {
		return err
	}

	retryWatcher, err := watchtools.NewRetryWatcher(pods.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector.String()
			return p.k8sClient.CoreV1().Pods(metav1.NamespaceAll).Watch(ctx, options)
		},
	})
	if err != nil {
		return fmt.Errorf("create retry watcher error: %v", err)
	}

	defer retryWatcher.Stop()
	logrus.Infof("start watching pods ......")

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context done, stop watching pods")
			return nil
		case e, ok := <-retryWatcher.ResultChan():
			if !ok {
				logrus.Warnf("pods retry watcher is closed")
				return nil
			}

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				logrus.Warnf("object is not a pod")
				continue
			}

			logrus.Debugf("watch pod event %s, object: %+v", e.Type, pod)

			switch e.Type {
			case watch.Added:
				addFunc(pod)
			case watch.Modified:
				modifyFunc(pod)
			case watch.Deleted:
				delFunc(pod)
			case watch.Bookmark, watch.Error:
				logrus.Debugf("ignore event %s, name: %s", e.Type, pod.Name)
			}
		}
	}
}

func (p *Pod) ListAllNamespace(fieldSelectors []string) (*corev1.PodList, error) {
	selector, err := fields.ParseSelector(strutil.Join(fieldSelectors, ","))
	if err != nil {
		return nil, err
	}

	return p.k8sClient.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}

func (p *Pod) ListNamespacePods(namespace string) (*corev1.PodList, error) {
	return p.k8sClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
}

type UnreadyReason string

const (
	None                  UnreadyReason = "None"
	ImagePullFailed       UnreadyReason = "ImagePullFailed"
	InsufficientResources UnreadyReason = "InsufficientResources"
	Unschedulable         UnreadyReason = "Unschedulable"
	ProbeFailed           UnreadyReason = "ProbeFailed"
	ContainerCannotRun    UnreadyReason = "ContainerCannotRun"
)

func (p *Pod) UnreadyPodReason(pod *corev1.Pod) (UnreadyReason, string) {
	// image pull failed
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil && container.State.Waiting.Reason == "ImagePullBackOff" {
			return ImagePullFailed, container.State.Waiting.Message
		}
	}

	// insufficient resources
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "PodScheduled" &&
			cond.Status == corev1.ConditionFalse &&
			cond.Reason == "Unschedulable" &&
			strutil.Contains(cond.Message, "nodes are available") &&
			strutil.Contains(cond.Message, "Insufficient memory", "Insufficient cpu") {
			return InsufficientResources, cond.Message
		}
	}

	// other unschedule reasons
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "PodScheduled" && cond.Status == corev1.ConditionFalse && cond.Reason == "Unschedulable" {
			return Unschedulable, cond.Message
		}
	}

	// container cannot run
	if pod.Status.Phase == "Running" {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == "ContainersReady" && cond.Status == corev1.ConditionFalse && cond.Reason == "ContainersNotReady" {
				for _, container := range pod.Status.ContainerStatuses {
					if container.Ready == false &&
						container.Started != nil &&
						*container.Started == false &&
						container.LastTerminationState.Terminated != nil &&
						container.LastTerminationState.Terminated.Reason == "ContainerCannotRun" {
						return ContainerCannotRun, container.LastTerminationState.Terminated.Message
					}
				}
			}
		}
	}
	// probe failed
	if pod.Status.Phase == "Running" {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == "ContainersReady" && cond.Status == corev1.ConditionFalse && cond.Reason == "ContainersNotReady" {
				return ProbeFailed, cond.Message
			}
		}
	}

	return None, ""
}

type PodStatus struct {
	Reason  UnreadyReason
	Message string
}

func (p *Pod) GetNamespacedPodsStatus(pods []corev1.Pod, serviceName string) ([]PodStatus, error) {
	r := make([]PodStatus, 0, len(pods))
	for _, pod := range pods {
		if serviceName == "" || !strings.Contains(pod.Name, serviceName) {
			continue
		}
		reason, message := p.UnreadyPodReason(&pod)
		switch reason {
		case None:
		case ImagePullFailed:
			r = append(r, PodStatus{Reason: ImagePullFailed, Message: message})
		case InsufficientResources:
			r = append(r, PodStatus{Reason: InsufficientResources, Message: message})
		case Unschedulable:
			r = append(r, PodStatus{Reason: Unschedulable, Message: message})
		case ProbeFailed:
			r = append(r, PodStatus{Reason: ProbeFailed, Message: message})
		case ContainerCannotRun:
			r = append(r, PodStatus{Reason: ContainerCannotRun, Message: message})
		}
	}
	return r, nil
}
