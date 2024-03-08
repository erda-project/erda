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

// Package deployment manipulates the k8s api of deployment object
package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/types"
)

// Deployment is the object to manipulate k8s api of deployment
type Deployment struct {
	cs kubernetes.Interface
}

// Option configures a Deployment
type Option func(*Deployment)

// New news a Deployment
func New(options ...Option) *Deployment {
	ns := &Deployment{}

	for _, op := range options {
		op(ns)
	}

	return ns
}

// WithClientSet with kubernetes clientSet
func WithClientSet(c kubernetes.Interface) Option {
	return func(d *Deployment) {
		d.cs = c
	}
}

// Patch  the k8s deployment object
func (d *Deployment) Patch(namespace, deploymentName, containerName string, snippet corev1.Container) error {
	// patch container with kubernetes snippet
	snippet.Name = containerName

	spec := types.PatchStruct{
		Spec: types.Spec{
			Template: types.PodTemplateSpec{
				Spec: types.PodSpec{
					Containers: []corev1.Container{
						snippet,
					},
				},
			},
		},
	}

	pathData, err := json.Marshal(spec)
	if err != nil {
		return errors.Errorf("failed to marshal patch data, %v", err)
	}

	if _, err := d.cs.AppsV1().Deployments(namespace).Patch(context.Background(), deploymentName,
		k8stypes.StrategicMergePatchType, pathData, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

// Create creates a k8s deployment object
func (d *Deployment) Create(deploy *appsv1.Deployment) error {
	if _, err := d.cs.AppsV1().Deployments(deploy.Namespace).
		Create(context.Background(), deploy, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

// Get gets a k8s deployment object
func (d *Deployment) Get(namespace, name string) (*appsv1.Deployment, error) {
	deploy, err := d.cs.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, k8serror.ErrNotFound
		}
		return nil, err
	}

	return deploy, nil
}

// List lists deployments under specific namespace
func (d *Deployment) List(namespace string, labelSelector map[string]string) (*appsv1.DeploymentList, error) {
	options := metav1.ListOptions{}

	if len(labelSelector) > 0 {
		kvs := make([]string, 0, len(labelSelector))
		for key, value := range labelSelector {
			kvs = append(kvs, fmt.Sprintf("%s=%s", key, value))
		}

		selector, err := labels.Parse(strings.Join(kvs, ","))
		if err != nil {
			return nil, errors.Errorf("failed to parse label selector, %v", err)
		}

		options.LabelSelector = selector.String()
	}

	deployList, err := d.cs.AppsV1().Deployments(namespace).List(context.Background(), options)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, k8serror.ErrNotFound
		}
		return nil, err
	}

	return deployList, nil
}

// Put updates a k8s deployment
func (d *Deployment) Put(deployment *appsv1.Deployment) error {
	if _, err := d.cs.AppsV1().Deployments(deployment.Namespace).
		Update(context.Background(), deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

// Delete deletes a k8s deployment
func (d *Deployment) Delete(namespace, name string) error {
	if err := d.cs.AppsV1().Deployments(namespace).
		Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return k8serror.ErrNotFound
		}
		return err
	}

	return nil
}

func (d *Deployment) WatchAllNamespace(ctx context.Context, addFunc, updateFunc, deleteFunc func(*appsv1.Deployment)) error {
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceSystem))
	if err != nil {
		return errors.Errorf("parse field selector error, %v", err)
	}

	deployments, err := d.cs.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}

	retryWatcher, err := watchtools.NewRetryWatcher(deployments.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector.String()
			return d.cs.AppsV1().Deployments(metav1.NamespaceAll).Watch(ctx, options)
		},
	})
	if err != nil {
		return fmt.Errorf("create retry watcher error: %v", err)
	}

	defer retryWatcher.Stop()
	logrus.Infof("start watching deployment ......")

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context done, stop watching deployments")
			return nil
		case e, ok := <-retryWatcher.ResultChan():
			if !ok {
				logrus.Warnf("pods retry watcher is closed")
				return nil
			}

			deploy, ok := e.Object.(*appsv1.Deployment)
			if !ok {
				logrus.Warnf("object is not a deployment")
				continue
			}

			logrus.Debugf("watch deployment, type: %s, object: %+v", e.Type, deploy)

			switch e.Type {
			case watch.Added:
				addFunc(deploy)
			case watch.Modified:
				updateFunc(deploy)
			case watch.Deleted:
				deleteFunc(deploy)
			case watch.Bookmark, watch.Error:
				logrus.Debugf("ignore event %s, name: %s", e.Type, deploy.Name)
			}
		}
	}
}

func (d *Deployment) LimitedListAllNamespace(limit int, cont *string) (*appsv1.DeploymentList, *string, error) {
	options := metav1.ListOptions{
		Limit: int64(limit),
	}

	// add continue
	if cont != nil {
		options.Continue = *cont
	}

	// parse field selector
	defaultSelector := []string{
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceDefault),
		fmt.Sprintf("metadata.namespace!=%s", metav1.NamespaceSystem),
	}
	fieldSelector, err := fields.ParseSelector(strings.Join(defaultSelector, ","))
	if err != nil {
		return nil, nil, err
	}
	options.FieldSelector = fieldSelector.String()

	// list with options
	deployList, err := d.cs.AppsV1().Deployments(metav1.NamespaceNone).List(context.Background(), options)
	if err != nil {
		return nil, nil, err
	}

	if deployList.ListMeta.Continue != "" {
		return deployList, &deployList.ListMeta.Continue, nil
	}

	return deployList, nil, nil
}
