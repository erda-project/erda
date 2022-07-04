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
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	kedav1alpha1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/scaledobject/keda/v1alpha1"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// ErdaScaledObject is the object to manipulate k8s crd api of scaledobject
type ErdaScaledObject struct {
	addr   string
	client *httpclient.HTTPClient
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
