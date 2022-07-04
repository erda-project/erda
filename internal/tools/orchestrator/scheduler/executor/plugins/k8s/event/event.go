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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// Event is the object to encapsulate docker
type Event struct {
	addr      string
	client    *httpclient.HTTPClient
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

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(e *Event) {
		e.addr = addr
		e.client = client
	}
}

func WithKubernetesClient(client kubernetes.Interface) Option {
	return func(event *Event) {
		event.k8sClient = client
	}
}

type rawevent struct {
	Type   string          `json:"type"`
	Object json.RawMessage `json:"object"`
}

func (e *Event) WatchPodEventsAllNamespaces(ctx context.Context, callback func(*apiv1.Event)) error {
	for {
		var b bytes.Buffer
		resp, err := e.client.Get(e.addr).
			Path("/api/v1/events").
			Param("limit", "10").
			Do().Body(&b)
		if err != nil {
			return err
		}
		if !resp.IsOK() {
			content, _ := ioutil.ReadAll(&b)
			errMsg := fmt.Sprintf("failed to get resp from k8s pods watcher, resp is not OK, body: %v",
				string(content))
			logrus.Errorf(errMsg)
			return errors.New(errMsg)
		}
		eventlist := apiv1.EventList{}
		if err := json.NewDecoder(&b).Decode(&eventlist); err != nil {
			return err
		}
		lastResourceVersion := eventlist.ListMeta.ResourceVersion
		body, resp, err := e.client.Get(e.addr).
			Path("/api/v1/watch/events").
			Header("Portal-SSE", "on").
			Param("fieldSelector", strutil.Join([]string{
				"metadata.namespace!=kube-system",
				"metadata.namespace!=kube-public",
				"involvedObject.kind=Pod",
			}, ",")).
			Param("resourceVersion", lastResourceVersion).
			Do().
			StreamBody()

		if err != nil {
			logrus.Errorf("failed to get resp from k8s events watcher, (%v)", err)
			return err
		}
		if !resp.IsOK() {
			errMsg := fmt.Sprintf("failed to get resp from k8s events watcher, resp is not OK")
			logrus.Errorf(errMsg)
			return errors.New(errMsg)
		}

		logrus.Info("get resp from k8s events watcher POD OK")
		decoder := json.NewDecoder(body)
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			raw := rawevent{}
			if err := decoder.Decode(&raw); err != nil {
				logrus.Errorf("failed to decode k8s event: %v", err)
				body.Close()
				break
			}
			ke := apiv1.Event{}
			if err := json.Unmarshal(raw.Object, &ke); err != nil {
				logrus.Errorf("failed to unmarshal k8s event obj, err: %v, raw: %s", err, string(raw.Object))
				body.Close()
				continue
			}
			callback(&ke)

		}
		if body != nil {
			body.Close()
		}
	}
}

// LimitedListAllNamespace limit list event all namespaces
func (e *Event) LimitedListAllNamespace(limit int, cont *string) (*apiv1.EventList, *string, error) {
	var eventlist apiv1.EventList
	var b bytes.Buffer
	req := e.client.Get(e.addr).
		Path("/api/v1/events")
	if cont != nil {
		req = req.Param("continue", *cont)
	}
	resp, err := req.Param("fieldSelector", strutil.Join([]string{
		"metadata.namespace!=default",
		"metadata.namespace!=kube-system",
		"metadata.namespace!=kube-public",
	}, ",")).
		Param("limit", fmt.Sprintf("%d", limit)).
		Do().Body(&b)
	if err != nil {
		return &eventlist, nil, err
	}
	if !resp.IsOK() {
		return &eventlist, nil, fmt.Errorf("failed to list k8s events, statuscode: %v, body: %v",
			resp.StatusCode(), b.String())
	}
	if err := json.NewDecoder(&b).Decode(&eventlist); err != nil {
		return &eventlist, nil, err
	}

	if eventlist.ListMeta.Continue != "" {
		return &eventlist, &eventlist.ListMeta.Continue, nil
	}
	return &eventlist, nil, nil
}

// ListByNamespace list event by namespace
func (e *Event) ListByNamespace(namespace string) (*apiv1.EventList, error) {
	var eventlist apiv1.EventList
	var b bytes.Buffer
	resp, err := e.client.Get(e.addr).
		Path("/api/v1/namespaces/" + namespace + "/events").
		Do().
		Body(&b)
	if err != nil {
		return &eventlist, err
	}
	if !resp.IsOK() {
		return &eventlist, fmt.Errorf("failed to list k8s events, namespace: %s, statuscode: %v, body: %v",
			namespace, resp.StatusCode(), b.String())
	}
	if err := json.NewDecoder(&b).Decode(&eventlist); err != nil {
		return &eventlist, err
	}

	if eventlist.ListMeta.Continue != "" {
		return &eventlist, nil
	}
	return &eventlist, nil
}

func (e *Event) WatchHPAEventsAllNamespaces(ctx context.Context, callback func(*apiv1.Event)) error {
	eventSelector, _ := fields.ParseSelector("involvedObject.kind=HorizontalPodAutoscaler")
	events, err := e.k8sClient.CoreV1().Events("").List(ctx, metav1.ListOptions{
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
			return e.k8sClient.CoreV1().Events("").Watch(context.Background(), options)
		},
	})

	if err != nil {
		return fmt.Errorf("create retry watcher error: %v", err)
	}

	defer retryWatcher.Stop()
	logrus.Infof("Start watching HPA events ......")

	for {
		select {
		case <-ctx.Done():
			return nil
		case eventEvent, ok := <-retryWatcher.ResultChan():
			logrus.Infof("HPA EVENT Watch An HPA Events with type %v", eventEvent.Type)
			if !ok {
				return nil
			}
			switch hpaEvent := eventEvent.Object.(type) {
			case *apiv1.Event:
				callback(hpaEvent)
			default:
			}
		}
	}
}
