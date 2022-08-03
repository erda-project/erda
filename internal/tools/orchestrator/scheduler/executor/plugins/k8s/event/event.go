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

// Package event manipulates the k8s api of event object
package event

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/pkg/strutil"
)

// Event is the object to encapsulate docker
type Event struct {
	addr      string
	k8sClient kubernetes.Interface
}

// Option configures an Event
type Option func(*Event)

// New news an Event
func New(options ...Option) *Event {
	ns := &Event{}
	for _, op := range options {
		op(ns)
	}

	return ns
}

func WithKubernetesClient(client kubernetes.Interface) Option {
	return func(event *Event) {
		event.k8sClient = client
	}
}

func (e *Event) WatchPodEventsAllNamespaces(ctx context.Context, callback func(*corev1.Event)) error {
	selector := strutil.Join([]string{
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceSystem),
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespacePublic),
		"involvedObject.kind=Pod",
	}, ",")
	podSelector, err := fields.ParseSelector(selector)
	if err != nil {
		return err
	}

	eventList, err := e.k8sClient.CoreV1().Events(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: podSelector.String(),
		Limit:         10,
	})

	if err != nil {
		return err
	}

	retryWatcher, err := watchtools.NewRetryWatcher(eventList.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = podSelector.String()
			return e.k8sClient.CoreV1().Events(metav1.NamespaceAll).Watch(context.Background(), options)
		},
	})
	if err != nil {
		return fmt.Errorf("create retry watcher error: %v", err)
	}

	defer retryWatcher.Stop()
	logrus.Infof("start watching pod events ......")

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context done, stop watching pod events")
			return nil
		case eventEvent, ok := <-retryWatcher.ResultChan():
			if !ok {
				logrus.Warnf("pod event retry watcher is closed")
				return nil
			}
			switch podEvent := eventEvent.Object.(type) {
			case *corev1.Event:
				callback(podEvent)
			default:
			}
		}
	}
}

// ListByNamespace list event by namespace
func (e *Event) ListByNamespace(namespace string) (*corev1.EventList, error) {
	eventList, err := e.k8sClient.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return eventList, nil
}

func (e *Event) WatchHPAEventsAllNamespaces(ctx context.Context, callback func(*corev1.Event)) error {
	eventSelector, err := fields.ParseSelector("involvedObject.kind=HorizontalPodAutoscaler")
	if err != nil {
		return err
	}
	events, err := e.k8sClient.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: eventSelector.String(),
	})
	if err != nil {
		return err
	}

	// record events in 2 minutes before orchestrator running, in case lost events between orchestrator restart
	now := time.Now().Add(-2 * time.Minute)
	for _, ev := range events.Items {
		if ev.LastTimestamp.After(now) {
			callback(&ev)
		}
	}

	lastResourceVersion := events.ListMeta.ResourceVersion

	// create retry watcher
	retryWatcher, err := watchtools.NewRetryWatcher(lastResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = eventSelector.String()
			return e.k8sClient.CoreV1().Events(metav1.NamespaceAll).Watch(context.Background(), options)
		},
	})

	if err != nil {
		return fmt.Errorf("create retry watcher error: %v", err)
	}

	defer retryWatcher.Stop()
	logrus.Infof("start watching hpa events ......")

	for {
		select {
		case <-ctx.Done():
			return nil
		case eventEvent, ok := <-retryWatcher.ResultChan():
			logrus.Infof("HPA EVENT Watch An HPA Events with type %v", eventEvent.Type)
			if !ok {
				logrus.Warnf("HPA event retry watcher is closed")
				return nil
			}
			switch hpaEvent := eventEvent.Object.(type) {
			case *corev1.Event:
				callback(hpaEvent)
			default:
			}
		}
	}
}
