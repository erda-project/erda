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

package daemonset

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/types"
)

type Daemonset struct {
	cs kubernetes.Interface
}

type Option func(*Daemonset)

func New(options ...Option) *Daemonset {
	ds := &Daemonset{}
	for _, op := range options {
		op(ds)
	}
	return ds
}

// WithClientSet with kubernetes clientSet
func WithClientSet(c kubernetes.Interface) Option {
	return func(d *Daemonset) {
		d.cs = c
	}
}

func (d *Daemonset) Create(ds *appsv1.DaemonSet) error {
	_, err := d.cs.AppsV1().DaemonSets(ds.Namespace).Create(context.Background(), ds, metav1.CreateOptions{})
	return err
}

func (d *Daemonset) Get(namespace, name string) (*appsv1.DaemonSet, error) {
	daemonSet, err := d.cs.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, k8serror.ErrNotFound
		}
		return nil, err
	}
	return daemonSet, err
}

func (d *Daemonset) List(namespace string, labelSelector map[string]string) (*appsv1.DaemonSetList, error) {
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

	daemonSetList, err := d.cs.AppsV1().DaemonSets(namespace).List(context.Background(), options)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, k8serror.ErrNotFound
		}
		return nil, err
	}
	return daemonSetList, nil
}

func (d *Daemonset) Update(daemonset *appsv1.DaemonSet) error {
	_, err := d.cs.AppsV1().DaemonSets(daemonset.Namespace).Update(context.Background(), daemonset, metav1.UpdateOptions{})
	return err
}

func (d *Daemonset) Delete(namespace, name string) error {
	if err := d.cs.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return k8serror.ErrNotFound
		}
		return err
	}
	return nil
}

func (d *Daemonset) Patch(namespace string, daemonsetName string, containerName string, snippet corev1.Container) error {
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

	if _, err := d.cs.AppsV1().DaemonSets(namespace).Patch(context.Background(), daemonsetName,
		k8stypes.StrategicMergePatchType, pathData, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}
