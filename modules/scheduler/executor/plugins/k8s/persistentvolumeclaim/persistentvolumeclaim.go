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

// Package persistentvolumeclaim manipulates the k8s api of persistentvolumeclaim object
package persistentvolumeclaim

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// PersistentVolumeClaim is the object to manipulate k8s api of persistentVolumeClaim
type PersistentVolumeClaim struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures a PersistentVolumeClaim
type Option func(*PersistentVolumeClaim)

// New news a PersistentVolumeClaim
func New(options ...Option) *PersistentVolumeClaim {
	pvc := &PersistentVolumeClaim{}

	for _, op := range options {
		op(pvc)
	}

	return pvc
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(n *PersistentVolumeClaim) {
		n.addr = addr
		n.client = client
	}
}

// Create creates a k8s persistentVolumeClaim
func (p *PersistentVolumeClaim) Create(pvc *apiv1.PersistentVolumeClaim) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", pvc.Namespace, "/persistentvolumeclaims")

	resp, err := p.client.Post(p.addr).
		Path(path).
		JSONBody(pvc).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create pvc, name: %s, (%v)", pvc.Name, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to create pvc, name: %s, statuscode: %v, body: %v",
			pvc.Name, resp.StatusCode(), b.String())
	}
	return nil
}
func (p *PersistentVolumeClaim) CreateIfNotExists(pvc *apiv1.PersistentVolumeClaim) error {
	var getb bytes.Buffer
	getpath := fmt.Sprintf("/api/v1/namespaces/%s/persistentvolumeclaims/%s", pvc.Namespace, pvc.Name)

	resp, err := p.client.Get(p.addr).
		Path(getpath).
		Do().
		Body(&getb)
	if err != nil {
		return errors.Errorf("failed to get pvc, name: %s, (%v)", pvc.Name, err)
	}
	if resp.IsOK() {
		return nil
	}
	if !resp.IsNotfound() {
		return errors.Errorf("failed to get pvc, name: %s, statuscode: %v, body: %v",
			pvc.Name, resp.StatusCode(), getb.String())
	}
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", pvc.Namespace, "/persistentvolumeclaims")

	resp, err = p.client.Post(p.addr).
		Path(path).
		JSONBody(pvc).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create pvc, name: %s, (%v)", pvc.Name, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to create pvc, name: %s, statuscode: %v, body: %v",
			pvc.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// Delete deletes a k8s persistentVolumeClaim
func (p *PersistentVolumeClaim) Delete(namespace, name string) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/persistentvolumeclaims/", name)

	resp, err := p.client.Delete(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete pvc, name: %s, (%v)", name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return k8serror.ErrNotFound
		}

		return errors.Errorf("failed to delete pvc, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}
