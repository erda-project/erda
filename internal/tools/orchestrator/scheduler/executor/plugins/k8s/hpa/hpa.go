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

package hpa

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// ErdaScaledObject is the object to manipulate k8s crd api of scaledobject
type ErdaHPA struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures a PersistentVolumeClaim
type Option func(*ErdaHPA)

// New news a PersistentVolumeClaim
func New(options ...Option) *ErdaHPA {
	hpa := &ErdaHPA{}

	for _, op := range options {
		op(hpa)
	}

	return hpa
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(n *ErdaHPA) {
		n.addr = addr
		n.client = client
	}
}

// Get gets a k8s hpa object
func (p *ErdaHPA) Get(namespace, name string) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	var b bytes.Buffer
	path := strutil.Concat("/apis/autoscaling/v2beta2/namespaces/", namespace, "/horizontalpodautoscalers/", name)

	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get hpa, namespace: %s name: %s", namespace, name)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, k8serror.ErrNotFound
		}
		return nil, errors.Errorf("failed to get hpa info, namespace: %s name: %s, statuscode: %v, body: %v", namespace, name, resp.StatusCode(), b.String())
	}

	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	if err := json.NewDecoder(&b).Decode(hpa); err != nil {
		return nil, err
	}
	return hpa, nil
}
