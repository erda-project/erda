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

// Package sparkapplication manipulates the k8s api of sparkapplication object
package sparkapplication

import (
	"bytes"
	"encoding/json"
	"fmt"

	sparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

var sparkAppLabelKey = "sparkoperator.k8s.io/app-name"

// SparkApplication is the object to manipulate k8s api of SparkApplication
type SparkApplication struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures a SparkApplication
type Option func(*SparkApplication)

// New news a SparkApplication
func New(options ...Option) *SparkApplication {
	ns := &SparkApplication{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(d *SparkApplication) {
		d.addr = addr
		d.client = client
	}
}

// Create creates a k8s SparkApplication object
func (s *SparkApplication) Create(app *sparkv1beta2.SparkApplication) error {
	var b bytes.Buffer

	resp, err := s.client.Post(s.addr).
		Path("/apis/sparkoperator.k8s.io/v1beta2/namespaces/" + app.Namespace + "/sparkapplications").
		JSONBody(app).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create spark application, name: %s, (%v)", app.Name, err)
	}

	if !resp.IsOK() {
		errMsg := fmt.Sprintf("failed to create spark application statuscode: %v, body: %v",
			resp.StatusCode(), b.String())
		return errors.Errorf(errMsg)
	}
	return nil
}

// Get gets a k8s SparkApplication object
func (s *SparkApplication) Get(namespace, name string) (*sparkv1beta2.SparkApplication, error) {
	var b bytes.Buffer
	resp, err := s.client.Get(s.addr).
		Path("/apis/sparkoperator.k8s.io/v1beta2/namespaces/" + namespace + "/sparkapplications/" + name).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("failed to get spark application info, name: %s, (%v)", name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, k8serror.ErrNotFound
		}
		return nil, errors.Errorf("failed to get spark application info, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}

	app := &sparkv1beta2.SparkApplication{}
	if err := json.NewDecoder(&b).Decode(app); err != nil {
		return nil, err
	}
	return app, nil
}

// List lists SparkApplication under specific namespace
func (s *SparkApplication) List(namespace string) (sparkv1beta2.SparkApplicationList, error) {
	var appList sparkv1beta2.SparkApplicationList

	var b bytes.Buffer
	resp, err := s.client.Get(s.addr).
		Path("/apis/sparkoperator.k8s.io/v1beta2/namespaces/" + namespace + "/sparkapplications").
		Do().
		Body(&b)

	if err != nil {
		return appList, errors.Errorf("failed to get spark application list, ns: %s, (%v)", namespace, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return appList, k8serror.ErrNotFound
		}
		return appList, errors.Errorf("failed to get spark application list, ns: %s, statuscode: %v, body: %v",
			namespace, resp.StatusCode(), b.String())
	}

	if err := json.NewDecoder(&b).Decode(&appList); err != nil {
		return appList, err
	}
	return appList, nil
}

// Update updates a k8s SparkApplication
func (s *SparkApplication) Update(app *sparkv1beta2.SparkApplication) error {
	var b bytes.Buffer

	// FIXME: spark operator 在 ResourceVersion 不变的情况下，不进行更新
	oriApp, err := s.Get(app.Namespace, app.Name)
	if err == nil {
		app.ObjectMeta.ResourceVersion = oriApp.ObjectMeta.ResourceVersion
	}

	resp, err := s.client.Put(s.addr).
		Path("/apis/sparkoperator.k8s.io/v1beta2/namespaces/" + app.Namespace + "/sparkapplications/" + app.Name).
		JSONBody(app).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to put spark application, name: %s, (%v)", app.Name, err)
	}

	if !resp.IsOK() {
		return errors.Errorf("failed to put spark application, name: %s, statuscode: %v, body: %v",
			app.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// DeletePod deletes the spark driver & executor pods by labelSelector
func (s *SparkApplication) DeletePod(namespace, name string) error {
	var b bytes.Buffer

	labelSelector := strutil.Concat(sparkAppLabelKey, "=", name)
	path := strutil.Concat("/api/v1/namespaces/", namespace, "/pods")

	resp, err := s.client.Delete(s.addr).
		Path(path).
		Param("labelSelector", labelSelector).
		Header("Content-Type", "application/json").
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete spark executor pods, namespace: %s, name: %s, (%v)",
			namespace, name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil
		}
		return errors.Errorf("failed to delete spark executor pods, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}

// Delete deletes a k8s SparkApplication
func (s *SparkApplication) Delete(namespace, name string) error {
	var b bytes.Buffer
	resp, err := s.client.Delete(s.addr).
		Path("/apis/sparkoperator.k8s.io/v1beta2/namespaces/" + namespace + "/sparkapplications/" + name).
		JSONBody(k8sapi.DeleteOptions).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete spark application, name: %s, (%v)", name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return k8serror.ErrNotFound
		}
		return errors.Errorf("failed to delete spark application, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}

// CreateOrUpdate create or update a k8s SparkApplication object
func (s *SparkApplication) CreateOrUpdate(app *sparkv1beta2.SparkApplication) error {
	var getErr error

	_, getErr = s.Get(app.Namespace, app.Name)
	if getErr == k8serror.ErrNotFound {
		return s.Create(app)
	}
	return s.Update(app)
}

// DeleteIfExists delete if k8s SparkApplication exists
func (s *SparkApplication) DeleteIfExists(namespace, name string) error {
	if err := s.Delete(namespace, name); err != nil {
		if err == k8serror.ErrNotFound {
			return nil
		}
		return err
	}

	return nil
}
