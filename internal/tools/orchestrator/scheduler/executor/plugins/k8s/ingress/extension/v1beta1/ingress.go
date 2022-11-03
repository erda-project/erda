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

package v1

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/ingress/common"
)

type Ingress struct {
	c v1beta1.ExtensionsV1beta1Interface
}

func NewIngress(c v1beta1.ExtensionsV1beta1Interface) *Ingress {
	return &Ingress{c: c}
}

func (i *Ingress) CreateIfNotExists(svc *apistructs.Service) error {
	if svc == nil {
		return errors.New("service is nil")
	}

	ing, err := buildIngress(svc)
	if err != nil || ing == nil {
		return err
	}

	if _, err = i.c.Ingresses(ing.Namespace).Get(context.Background(), ing.Name, metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		_, err = i.c.Ingresses(ing.Namespace).Create(context.Background(), ing, metav1.CreateOptions{})
		return err
	}

	logrus.Warnf("ingress %s in namespace %s already existed", svc.Name, svc.Namespace)

	return nil
}

func buildIngress(svc *apistructs.Service) (*extensionsv1beta1.Ingress, error) {
	publicHosts := common.ParsePublicHostsFromLabel(svc.Labels)
	if len(publicHosts) == 0 {
		return nil, nil
	}

	// create ingress
	rules := buildRules(publicHosts, svc.Name, svc.Ports[0].Port)
	tls := buildTLS(publicHosts)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: rules,
			TLS:   tls,
		},
	}, nil
}

func buildRules(publicHosts []string, name string, port int) []extensionsv1beta1.IngressRule {
	rules := make([]extensionsv1beta1.IngressRule, len(publicHosts))
	for i, host := range publicHosts {
		rules[i].Host = host
		rules[i].HTTP = &extensionsv1beta1.HTTPIngressRuleValue{
			Paths: []extensionsv1beta1.HTTPIngressPath{
				{
					//TODO: add Path
					// Path:
					Backend: extensionsv1beta1.IngressBackend{
						ServiceName: name,
						ServicePort: intstr.FromInt(port),
					},
				},
			},
		}
	}
	return rules
}

func buildTLS(publicHosts []string) []extensionsv1beta1.IngressTLS {
	tls := make([]extensionsv1beta1.IngressTLS, 1)
	tls[0].Hosts = make([]string, len(publicHosts))
	for i, host := range publicHosts {
		tls[0].Hosts[i] = host
	}
	return tls
}
