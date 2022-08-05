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

// Package namespace manipulates the k8s api of namespace object
package namespace

import (
	"context"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
)

// Namespace is the object to manipulate k8s api of namespace
type Namespace struct {
	cs kubernetes.Interface
}

// Option configures a Namespace
type Option func(*Namespace)

// New news a Namespace
func New(options ...Option) *Namespace {
	ns := &Namespace{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

// WithKubernetesClient provides an Option
func WithKubernetesClient(k kubernetes.Interface) Option {
	return func(n *Namespace) {
		n.cs = k
	}
}

// Create creates a k8s namespace
// TODO: Need to pass in the namespace structure
func (n *Namespace) Create(ns string, labels map[string]string) error {
	if _, err := n.cs.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
			Labels: labels,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}
	logrus.Infof("succeed to create namespace %s, labels %+v", ns, labels)
	return nil
}

func (n *Namespace) Update(ns string, labels map[string]string) error {
	if _, err := n.cs.CoreV1().Namespaces().Update(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
			Labels: labels,
		},
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}

	logrus.Infof("succeed to update namespace %s, labels: %+v", ns, labels)
	return nil
}

// Exists decides whether a namespace exists
func (n *Namespace) Exists(ns string) error {
	_, err := n.cs.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return k8serror.ErrNotFound
		}
		return err
	}
	return nil
}

// Delete deletes a k8s namespace (deletes all dependents in the foreground)
func (n *Namespace) Delete(ns string, force ...bool) error {
	deleteOptions := metav1.DeleteOptions{}
	if len(force) != 0 && force[0] {
		propagationPolicy := metav1.DeletePropagationForeground
		deleteOptions.PropagationPolicy = &propagationPolicy
		logrus.Debugf("force delete namespace %s", ns)
	}

	err := n.cs.CoreV1().Namespaces().Delete(context.Background(), ns, deleteOptions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return k8serror.ErrNotFound
		}
		return err
	}
	return nil
}
