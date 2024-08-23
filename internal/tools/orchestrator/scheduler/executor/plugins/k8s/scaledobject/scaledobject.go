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

package scaledobject

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	vpatypes "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// ErdaScaledObject is the object to manipulate k8s crd api of scaledobject
type ErdaScaledObject struct {
	addr      string
	client    *httpclient.HTTPClient
	vpaClient *vpa_clientset.Clientset
}

// Option configures a PersistentVolumeClaim
type Option func(*ErdaScaledObject)

// New news a PersistentVolumeClaim
func New(options ...Option) *ErdaScaledObject {
	scaledObj := &ErdaScaledObject{}

	for _, op := range options {
		op(scaledObj)
	}

	return scaledObj
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(n *ErdaScaledObject) {
		n.addr = addr
		n.client = client
	}
}

func WithVPAClient(client *vpa_clientset.Clientset) Option {
	return func(n *ErdaScaledObject) {
		n.vpaClient = client
	}
}

// Create creates a k8s keda crd scaledObject
func (p *ErdaScaledObject) Create(scaledObject *kedav1alpha1.ScaledObject) error {
	var b bytes.Buffer
	path := strutil.Concat("/apis/keda.sh/v1alpha1/namespaces/", scaledObject.Namespace, "/scaledobjects")

	resp, err := p.client.Post(p.addr).
		Path(path).
		JSONBody(scaledObject).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create scaledobject in namespace %s with name %s, error: %v", scaledObject.Namespace, scaledObject.Name, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to create scaledobject in namespace %s with name %s, statuscode: %v, body: %v",
			scaledObject.Namespace, scaledObject.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// Delete deletes a k8s keda crd scaledObject
func (p *ErdaScaledObject) Delete(namespace, name string) error {
	var b bytes.Buffer
	path := strutil.Concat("/apis/keda.sh/v1alpha1/namespaces/", namespace, "/scaledobjects/", name)

	resp, err := p.client.Delete(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete scaledobject in namespace %s with name %s, error: %v", namespace, name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return k8serror.ErrNotFound
		}

		return errors.Errorf("failed to delete scaledobject in namespace %s with name %s, statuscode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}
	return nil
}

// Get gets a k8s keda crd scaledObject
func (p *ErdaScaledObject) Get(namespace, name string) (*kedav1alpha1.ScaledObject, error) {
	var b bytes.Buffer
	path := strutil.Concat("/apis/keda.sh/v1alpha1/namespaces/", namespace, "/scaledobjects/", name)

	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get scaledobject, namespace: %s name: %s", namespace, name)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, k8serror.ErrNotFound
		}
		return nil, errors.Errorf("failed to get scaledobject info, namespace: %s name: %s, statuscode: %v, body: %v", namespace, name, resp.StatusCode(), b.String())
	}

	scaledObject := &kedav1alpha1.ScaledObject{}
	if err := json.NewDecoder(&b).Decode(scaledObject); err != nil {
		return nil, err
	}
	return scaledObject, nil
}

type PatchStruct struct {
	Spec kedav1alpha1.ScaledObjectSpec `json:"spec"`
}

// Patch patchs the k8s keda crd scaledObject object
func (p *ErdaScaledObject) Patch(namespace, name string, patch *kedav1alpha1.ScaledObject) error {
	spec := PatchStruct{}

	spec.Spec.ScaleTargetRef = patch.Spec.ScaleTargetRef
	if *patch.Spec.MinReplicaCount > 0 && *patch.Spec.MaxReplicaCount > 0 && *patch.Spec.MinReplicaCount < *patch.Spec.MaxReplicaCount {
		spec.Spec.MinReplicaCount = patch.Spec.MinReplicaCount
		spec.Spec.MaxReplicaCount = patch.Spec.MaxReplicaCount
	}

	if len(patch.Spec.Triggers) > 0 {
		spec.Spec.Triggers = patch.Spec.Triggers
	}

	var b bytes.Buffer
	path := strutil.Concat("/apis/keda.sh/v1alpha1/namespaces/", namespace, "/scaledobjects/", name)
	resp, err := p.client.Patch(p.addr).
		Path(path).
		JSONBody(spec).
		Header("Content-Type", "application/merge-patch+json").
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to patch scaledobject, namespace: %s name: %s, error: %v", namespace, name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to patch scaledobject, namespace: %s name: %s, statuscode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}
	return nil
}

// CreateVPA creates a k8s vpa object
func (p *ErdaScaledObject) CreateVPA(scaledObject *vpatypes.VerticalPodAutoscaler) error {
	var b bytes.Buffer
	// /apis/autoscaling.k8s.io/v1/namespaces/default/verticalpodautoscalers?fieldManager=kubectl-create 201 Created in 9 milliseconds
	path := strutil.Concat("/apis/autoscaling.k8s.io/v1/namespaces/", scaledObject.Namespace, "/verticalpodautoscalers")

	resp, err := p.client.Post(p.addr).
		Path(path).
		JSONBody(scaledObject).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create vpa in namespace %s with name %s, error: %v", scaledObject.Namespace, scaledObject.Name, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to create vpa in namespace %s with name %s, statuscode: %v, body: %v",
			scaledObject.Namespace, scaledObject.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// DeleteVPA deletes a k8s vpa
func (p *ErdaScaledObject) DeleteVPA(namespace, name string) error {
	var b bytes.Buffer
	path := strutil.Concat("/apis/autoscaling.k8s.io/v1/namespaces/", namespace, "/verticalpodautoscalers/", name)

	resp, err := p.client.Delete(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete scaledobject in namespace %s with name %s, error: %v", namespace, name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return k8serror.ErrNotFound
		}

		return errors.Errorf("failed to delete scaledobject in namespace %s with name %s, statuscode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}
	return nil
}

// GetVPA gets a k8s vpa
func (p *ErdaScaledObject) GetVPA(namespace, name string) (*vpatypes.VerticalPodAutoscaler, error) {
	var b bytes.Buffer
	path := strutil.Concat("/apis/autoscaling.k8s.io/v1/namespaces/", namespace, "/verticalpodautoscalers/", name)

	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get vpa, namespace: %s name: %s", namespace, name)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, k8serror.ErrNotFound
		}
		return nil, errors.Errorf("failed to get vpa info, namespace: %s name: %s, statuscode: %v, body: %v", namespace, name, resp.StatusCode(), b.String())
	}

	scaledObject := &vpatypes.VerticalPodAutoscaler{}
	if err := json.NewDecoder(&b).Decode(scaledObject); err != nil {
		return nil, err
	}
	return scaledObject, nil
}

type PatchVPAStruct struct {
	Spec vpatypes.VerticalPodAutoscalerSpec `json:"spec"`
}

// PatchVPA patchs the k8s vpa object
func (p *ErdaScaledObject) PatchVPA(namespace, name string, patch *vpatypes.VerticalPodAutoscaler) error {
	spec := PatchVPAStruct{}

	spec.Spec.TargetRef = patch.Spec.TargetRef
	if patch.Spec.UpdatePolicy != nil && patch.Spec.UpdatePolicy.UpdateMode != nil {
		spec.Spec.UpdatePolicy = patch.Spec.UpdatePolicy
	}
	if patch.Spec.ResourcePolicy != nil && len(patch.Spec.ResourcePolicy.ContainerPolicies) > 0 {
		spec.Spec.ResourcePolicy = patch.Spec.ResourcePolicy
	}

	var b bytes.Buffer
	path := strutil.Concat("/apis/autoscaling.k8s.io/v1/namespaces/", namespace, "/verticalpodautoscalers/", name)
	resp, err := p.client.Patch(p.addr).
		Path(path).
		JSONBody(spec).
		Header("Content-Type", "application/merge-patch+json").
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to patch vpa, namespace: %s name: %s, error: %v", namespace, name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to patch vpa, namespace: %s name: %s, statuscode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}
	return nil
}

func (p *ErdaScaledObject) WatchVPAsAllNamespaces(ctx context.Context, callback func(*vpatypes.VerticalPodAutoscaler)) error {
	selector := strutil.Join([]string{
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceSystem),
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespacePublic),
	}, ",")
	podSelector, err := fields.ParseSelector(selector)
	if err != nil {
		return err
	}

	vpaList, err := p.vpaClient.AutoscalingV1().VerticalPodAutoscalers(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: podSelector.String(),
		Limit:         10,
	})

	if err != nil {
		return err
	}

	retryWatcher, err := watchtools.NewRetryWatcher(vpaList.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = podSelector.String()
			return p.vpaClient.AutoscalingV1().VerticalPodAutoscalers(corev1.NamespaceAll).Watch(context.Background(), options)
		},
	})

	defer retryWatcher.Stop()
	logrus.Infof("start watching vpa ......")

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context done, stop watching vpa")
			return nil
		case vpa, ok := <-retryWatcher.ResultChan():
			if !ok {
				logrus.Warnf("vpa retry watcher is closed")
				return nil
			}
			switch podEvent := vpa.Object.(type) {
			case *vpatypes.VerticalPodAutoscaler:
				callback(podEvent)
			default:
			}
		}
	}
}
