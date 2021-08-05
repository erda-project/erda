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

package k8s

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateService create k8s service
func (k *Kubernetes) CreateService(service *apistructs.Service, selectors map[string]string) error {
	svc := newService(service, selectors)
	return k.service.Create(svc)
}

// PutService update k8s service
func (k *Kubernetes) PutService(svc *apiv1.Service) error {
	return k.service.Put(svc)
}

// GetService get k8s service
func (k *Kubernetes) GetService(namespace, name string) (*apiv1.Service, error) {
	return k.service.Get(namespace, name)
}

// DeleteService delete k8s service
func (k *Kubernetes) DeleteService(namespace, name string) error {
	return k.service.Delete(namespace, name)
}

func (k *Kubernetes) ListService(namespace string, selectors map[string]string) (apiv1.ServiceList, error) {
	return k.service.List(namespace, selectors)
}

// Port changes in the service description will cause changes in services and ingress
func (k *Kubernetes) updateService(service *apistructs.Service, selectors map[string]string) error {
	// Service.Ports is empty, indicating that no service is expected
	if len(service.Ports) == 0 {
		// There is a service before the update, if there is no service, delete the servicece
		if err := k.DeleteService(service.Namespace, service.Name); err != nil {
			return err
		}

		return nil
	}

	svc, getErr := k.GetService(service.Namespace, service.Name)
	if getErr != nil && getErr != k8serror.ErrNotFound {
		return errors.Errorf("failed to get service, name: %s, (%v)", service.Name, getErr)
	}

	// If not found, create a new k8s service
	if getErr == k8serror.ErrNotFound {
		if err := k.CreateService(service, selectors); err != nil {
			return err
		}

	} else {
		if err := k.UpdateK8sService(svc, service, selectors); err != nil {
			return err
		}
	}

	return nil
}

func newService(service *apistructs.Service, selectors map[string]string) *apiv1.Service {
	if len(service.Ports) == 0 {
		return nil
	}

	k8sService := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels:    make(map[string]string),
		},
		Spec: apiv1.ServiceSpec{
			// TODO: type?
			//Type: ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"app": service.Name,
			},
		},
	}

	setServiceLabelSelector(k8sService, selectors)

	for i, port := range service.Ports {
		k8sService.Spec.Ports = append(k8sService.Spec.Ports, apiv1.ServicePort{
			// TODO: name?
			Name: strutil.Concat(strings.ToLower(port.Protocol), "-", strconv.Itoa(i)),
			Port: int32(port.Port),
			// The user on Dice only fills in Port, that is, Port (port exposed by service) and targetPort (port exposed by container) are the same
			TargetPort: intstr.FromInt(port.Port),
			// Append protocol feature, Protocol Type contains TCP, UDP, SCTP
			Protocol: port.L4Protocol,
		})
	}
	return k8sService
}

// TODO: Complete me.
func diffServiceMetadata(left, right *apiv1.Service) bool {
	// compare the fields of Metadata and Spec
	if !reflect.DeepEqual(left.Labels, right.Labels) {
		return true
	}

	if !reflect.DeepEqual(left.Spec.Ports, right.Spec.Ports) {
		return true
	}

	if !reflect.DeepEqual(left.Spec.Selector, right.Spec.Selector) {
		return true
	}

	return false
}

// actually get deployment's names list, as k8s service would not be created
// if no ports exposed
func (k *Kubernetes) listServiceName(namespace string, labelSelector map[string]string) ([]string, error) {
	strs := make([]string, 0)
	deployList, err := k.deploy.List(namespace, labelSelector)
	if err != nil {
		return strs, err
	}

	for _, item := range deployList.Items {
		strs = append(strs, item.Name)
	}

	daemonSets, err := k.ds.List(namespace, labelSelector)
	if err != nil {
		return strs, err
	}
	for _, item := range daemonSets.Items {
		strs = append(strs, item.Name)
	}

	return strs, nil
}

func (k *Kubernetes) getRuntimeServiceMap(namespace string, labelSelector map[string]string) (map[string]RuntimeServiceOperator, error) {
	m := map[string]RuntimeServiceOperator{}
	deployList, err := k.deploy.List(namespace, labelSelector)
	if err != nil {
		return map[string]RuntimeServiceOperator{}, err
	}

	for _, item := range deployList.Items {
		m[item.Name] = RuntimeServiceDelete
	}

	daemonSets, err := k.ds.List(namespace, labelSelector)
	if err != nil {
		return map[string]RuntimeServiceOperator{}, err
	}
	for _, item := range daemonSets.Items {
		m[item.Name] = RuntimeServiceDelete
	}

	return m, nil
}

func getServiceName(service *apistructs.Service) string {
	if service.ProjectServiceName != "" {
		return service.ProjectServiceName
	}
	return service.Name
}

func setServiceLabelSelector(k8sService *apiv1.Service, selectors map[string]string) {
	for k, v := range selectors {
		k8sService.Spec.Selector[k] = v
	}
}

func (k *Kubernetes) deleteRuntimeServiceWithProjectNamespace(service apistructs.Service) error {
	if service.ProjectServiceName != "" {
		err := k.service.Delete(service.Namespace, service.ProjectServiceName)
		if err != nil {
			return fmt.Errorf("delete service %s error: %v", service.Name, err)
		}
	}
	if projectService, ok := deepcopy.Copy(service).(apistructs.Service); ok {
		projectService.Name = projectService.ProjectServiceName
		if k.istioEngine != istioctl.EmptyEngine {
			err := k.istioEngine.OnServiceOperator(istioctl.ServiceDelete, &projectService)
			if err != nil {
				return fmt.Errorf("delete istio resource error: %v", err)
			}
		}
	}
	return nil
}

func (k *Kubernetes) createProjectService(service *apistructs.Service, sgID string) error {
	if service.ProjectServiceName != "" {
		projectService, ok := deepcopy.Copy(service).(*apistructs.Service)
		if ok {
			selectors := map[string]string{
				LabelServiceGroupID: sgID,
				"app":               service.Name,
			}
			projectService.Name = service.ProjectServiceName
			projectServiceErr := k.CreateService(projectService, selectors)
			if projectServiceErr != nil {
				errMsg := fmt.Sprintf("Create project service %s err %v", projectService.Name, projectServiceErr)
				logrus.Errorf(errMsg)
				return fmt.Errorf(errMsg)
			}
		}
	}
	return nil
}

func (k *Kubernetes) deleteProjectService(service *apistructs.Service) error {
	if service.ProjectServiceName != "" {
		projectService, ok := deepcopy.Copy(service).(*apistructs.Service)
		if ok {
			projectService.Name = service.ProjectServiceName
			projectServiceErr := k.DeleteService(projectService.Namespace, projectService.Name)
			if projectServiceErr != nil {
				errMsg := fmt.Sprintf("Delete project service %s err %v", projectService.Name, projectServiceErr)
				logrus.Errorf(errMsg)
				return fmt.Errorf(errMsg)
			}
		}
	}
	logrus.Infof("delete the kubernetes service %s on namespace %s successfully", service.ProjectServiceName, service.Namespace)
	return nil
}

func (k *Kubernetes) updateProjectService(service *apistructs.Service, sgID string) error {
	if service.ProjectServiceName != "" {
		projectService, ok := deepcopy.Copy(service).(*apistructs.Service)
		if ok {
			selectors := map[string]string{
				LabelServiceGroupID: sgID,
				"app":               service.Name,
			}
			projectService.Name = service.ProjectServiceName

			projectServiceErr := k.updateService(projectService, selectors)
			if projectServiceErr != nil {
				errMsg := fmt.Sprintf("Update project service %s err %v", projectService.Name, projectServiceErr)
				logrus.Errorf(errMsg)
				return fmt.Errorf(errMsg)
			}
		}
	}
	return nil
}

func (k *Kubernetes) UpdateK8sService(k8sService *apiv1.Service, service *apistructs.Service, selectors map[string]string) error {
	newPorts := []apiv1.ServicePort{}
	for i, p := range service.Ports {
		port := apiv1.ServicePort{
			Name:       strutil.Concat(strings.ToLower(p.Protocol), "-", strconv.Itoa(i)),
			Port:       int32(p.Port),
			TargetPort: intstr.FromInt(p.Port),
			Protocol:   p.L4Protocol,
		}
		newPorts = append(newPorts, port)
	}

	setServiceLabelSelector(k8sService, selectors)
	k8sService.Spec.Ports = newPorts

	if err := k.PutService(k8sService); err != nil {
		errMsg := fmt.Sprintf("update service err %v", err)
		logrus.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}
