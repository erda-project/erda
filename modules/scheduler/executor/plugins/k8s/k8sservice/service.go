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

// Package k8sservice manipulates the k8s api of service object
package k8sservice

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// Service is the object to manipulate k8s api of service
type Service struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures a Service
type Option func(*Service)

// New news a Service
func New(options ...Option) *Service {
	ns := &Service{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(s *Service) {
		s.addr = addr
		s.client = client
	}
}

// Create creates a k8s service
func (s *Service) Create(service *apiv1.Service) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", service.Namespace, "/services")

	resp, err := s.client.Post(s.addr).
		Path(path).
		JSONBody(service).
		Do().
		Body(&b)

	if err != nil {
		return errors.Wrapf(err, "failed to create service, name: %s, (%v)", service.Name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to create service, name: %s, statuscode: %v, body: %v",
			service.Name, resp.StatusCode(), b.String())
	}
	return nil
}

func (s *Service) List(namespace string) (apiv1.ServiceList, error) {
	var b bytes.Buffer
	svclist := apiv1.ServiceList{}

	path := strutil.Concat("/api/v1/namespaces/", namespace, "/services")
	resp, err := s.client.Get(s.addr).
		Path(path).
		Do().
		Body(&b)
	if err != nil {
		return svclist, errors.Errorf("failed to get service list, namespace: %s, err: %v", namespace, err)
	}
	if !resp.IsOK() {
		return svclist, errors.Errorf("failed to get service list, namespace: %s, statuscode: %v, body: %v",
			namespace, resp.StatusCode(), b.String())
	}
	if err := json.NewDecoder(&b).Decode(&svclist); err != nil {
		return svclist, err
	}
	return svclist, nil
}

// Get gets a k8s service
func (s *Service) Get(namespace, name string) (*apiv1.Service, error) {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/services/", name)

	resp, err := s.client.Get(s.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get service, name: %s, (%v)", name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, k8serror.ErrNotFound
		}
		return nil, errors.Errorf("failed to get service, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}

	svc := &apiv1.Service{}
	if err := json.NewDecoder(&b).Decode(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

// Put updates a k8s service
func (s *Service) Put(service *apiv1.Service) error {
	var b bytes.Buffer
	namespace := service.Namespace
	name := service.Name
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/services/", name)

	resp, err := s.client.Put(s.addr).
		Path(path).
		JSONBody(service).
		Do().
		Body(&b)

	if err != nil {
		return errors.Wrapf(err, "failed to put service, name: %s, (%v)", name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to put service, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}

// Delete deletes a k8s service
func (s *Service) Delete(namespace, name string) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/services/", name)

	resp, err := s.client.Delete(s.addr).
		Path(path).
		JSONBody(k8sapi.DeleteOptions).
		Do().
		Body(&b)

	if err != nil {
		return errors.Wrapf(err, "failed to delete service, name: %s, (%v)", name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			logrus.Debugf("service not found, ns: %s, name: %s", namespace, name)
			return nil
		}
		return errors.Errorf("failed to delete service, name: %s, status code: %v, resp body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}
