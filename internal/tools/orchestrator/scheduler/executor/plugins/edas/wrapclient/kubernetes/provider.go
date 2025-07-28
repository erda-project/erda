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

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/apistructs"
)

// refactor, source: edas/edas.go
// PR: https://github.com/erda-project/erda/pull/6102

type Interface interface {
	GetK8sService(name string) (*corev1.Service, error)
	GetK8sDeployList(group string, services *[]apistructs.Service) error
	CreateK8sService(appName, sgID, serviceName string, ports []int) error
	CreateOrUpdateK8sService(ctx context.Context, appName, sgID string, serviceName string, ports []int) error
	DeleteK8sService(appName string) error
}

type wrapKubernetes struct {
	l *logrus.Entry

	namespace string
	cs        kubernetes.Interface
}

func New(l *logrus.Entry, cs kubernetes.Interface, namespace string) Interface {
	return &wrapKubernetes{
		l:         l.WithField("wrap-client", "kubernetes"),
		cs:        cs,
		namespace: namespace,
	}
}
