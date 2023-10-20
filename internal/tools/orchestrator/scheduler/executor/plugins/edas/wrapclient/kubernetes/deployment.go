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
	goerrors "errors"
	"strings"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// service name env of edas service
var diceServiceName = "DICE_SERVICE_NAME"

// GetK8sDeployList get the deployment list of corresponding runtime from k8s api
// TODO: use wrapEDASClient
func (e *wrapKubernetes) GetK8sDeployList(group string, services *[]apistructs.Service) error {
	var (
		edasAhasName string
		port         int32
		cpu          int64
		mem          int64
		replicas     int32
		image        string
	)

	l := e.l.WithField("func", "GetK8sDeployList")

	l.Debugf("get deploylist from group: %+v", group)

	deployList, err := e.cs.AppsV1().Deployments(e.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	l.Debugf("get deploylist of old runtime from k8s : %+v", deployList.Items)
	for _, i := range deployList.Items {
		// Get the deployed deployment of the runtime from the deploylist
		l.Debugf("deploy name: %+v", i.ObjectMeta.Name)
		if strings.Contains(i.ObjectMeta.Name, group) && *i.Spec.Replicas != 0 {
			var iService apistructs.Service
			for _, j := range i.Spec.Template.Spec.Containers[0].Env {
				if j.Name == diceServiceName {
					edasAhasName = j.Value
					cpu = i.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()
					mem, _ = i.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().AsInt64()
					replicas = *i.Spec.Replicas
					image = i.Spec.Template.Spec.Containers[0].Image
				}
			}
			// Query the service interface to get the port list
			appName := group + "-" + edasAhasName
			if kubeSvc, err := e.GetK8sService(appName); err != nil {
				if goerrors.Is(err, ServiceNotFound) {
					port = 0
				} else {
					return errors.Errorf("get k8s service err: %+v", err)
				}
			} else {
				port = kubeSvc.Spec.Ports[0].Port
			}
			iService.Name = edasAhasName
			iService.Ports = append(iService.Ports, diceyml.ServicePort{Port: int(port), Protocol: "TCP", L4Protocol: apiv1.ProtocolTCP})
			iService.Scale = int(replicas)
			iService.Resources.Cpu = float64(cpu)
			iService.Resources.Mem = float64(mem)
			iService.Image = image
			*services = append(*services, iService)
		}
	}

	l.Debugf("old service list : %+v", services)
	return nil
}
