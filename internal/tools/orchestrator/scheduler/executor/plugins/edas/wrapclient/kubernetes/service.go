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

package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetK8sService container service, k8s service dns record: svc.Name + ".default.svc.cluster.local"
// TODO: Implement the use of a selector when adding labels in the service.
// In certain legacy EDAS environments, there is an 'slbPrefix' configuration.
func (e *wrapKubernetes) GetK8sService(name string) (*corev1.Service, error) {
	if len(name) == 0 {
		return nil, errors.Errorf("get k8s service: invalid params")
	}

	slbPrefix := fmt.Sprintf("%s-%s", types.EDAS_SLB_INTERNAL, name)
	svcList, err := e.cs.
		CoreV1().Services(e.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "get k8s service, name: %s", name)
	}

	for _, svc := range svcList.Items {
		switch svc.Spec.Type {
		case corev1.ServiceTypeClusterIP:
			if strings.HasPrefix(svc.Name, name) {
				return &svc, nil
			}
		case corev1.ServiceTypeLoadBalancer:
			// TODO: remove edas basic selector
			if strings.Compare(svc.Spec.Selector["edas-domain"], "edas-admin") == 0 &&
				strings.HasPrefix(svc.Name, slbPrefix) {
				return &svc, nil
			}
		}
	}

	return nil, ServiceNotFound
}

// CreateK8sService create kubernetes service
func (e *wrapKubernetes) CreateK8sService(appName string, appID string, ports []int) error {
	l := e.l.WithField("func", "CreateK8sService")

	k8sService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appName,
			Labels: make(map[string]string),
		},
		Spec: corev1.ServiceSpec{
			// TODO: type?
			//Type: ServiceTypeLoadBalancer,
			Selector: make(map[string]string),
		},
	}
	k8sService.Spec.Selector["edas.appid"] = appID
	servicenamePrefix := "http-"
	for i, port := range ports {
		k8sService.Spec.Ports = append(k8sService.Spec.Ports, corev1.ServicePort{
			// TODO: name?
			Name:       strutil.Concat(servicenamePrefix, strconv.Itoa(i)),
			Port:       int32(port),
			TargetPort: intstr.FromInt(port),
		})
	}

	l.Infof("start to create k8s svc, appName: %s", appName)
	_, err := e.cs.CoreV1().Services(e.namespace).Create(context.TODO(), k8sService, metav1.CreateOptions{})
	return err
}

// CreateK8sServiceIfNotExist create kubernetes service if not exist
func (e *wrapKubernetes) CreateK8sServiceIfNotExist(appName string, appID string, ports []int) error {
	_, err := e.cs.CoreV1().Services(e.namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return e.CreateK8sService(appName, appID, ports)
		}
		return err
	}

	return nil
}

// DeleteK8sService delete kubernetes service
func (e *wrapKubernetes) DeleteK8sService(appName string) error {
	err := e.cs.CoreV1().Services(e.namespace).Delete(context.TODO(), appName, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}
