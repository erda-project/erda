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

package ingress

import (
	"github.com/sirupsen/logrus"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/apistructs"
	ingv1beta1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/ingress/extension/v1beta1"
	ingv1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/ingress/networking/v1"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
)

type Interface interface {
	CreateIfNotExists(svc *apistructs.Service) error
}

// New TODO: Instead of operate resource ingress to hepa component
func New(c kubernetes.Interface) (Interface, error) {
	ok, err := util.VersionHas(c, extensionsv1beta1.SchemeGroupVersion.String())
	if err != nil {
		return nil, err
	}
	if ok {
		logrus.Infof("ingress helper use version: %s", extensionsv1beta1.SchemeGroupVersion.String())
		return ingv1beta1.NewIngress(c.ExtensionsV1beta1()), nil
	}

	logrus.Infof("ingress helper use version: %s", networkingv1.SchemeGroupVersion.String())
	return ingv1.NewIngress(c.NetworkingV1()), nil
}
