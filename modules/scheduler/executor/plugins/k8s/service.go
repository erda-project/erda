package k8s

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/erda-project/erda/pkg/istioctl"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateService create k8s service
func (k *Kubernetes) CreateService(service *apistructs.Service) error {
	svc := newService(service)
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

// 服务描述中的端口变化会导致services及ingress的变化
func (k *Kubernetes) updateService(service *apistructs.Service) error {
	// service.Ports为空表明期望无service
	if len(service.Ports) == 0 {
		// 更新前有service, 期望无service, 则删除service
		if err := k.DeleteService(service.Namespace, service.Name); err != nil {
			return err
		}

		return nil
	}

	svc, getErr := k.GetService(service.Namespace, service.Name)
	if getErr != nil && getErr != k8serror.ErrNotFound {
		return errors.Errorf("failed to get service, name: %s, (%v)", service.Name, getErr)
	}

	// 如果没有找到，则新建 k8s service
	if getErr == k8serror.ErrNotFound {
		if err := k.CreateService(service); err != nil {
			return err
		}
	} else {
		// 更新前有service，期望有service，则更新service
		desiredService := newService(service)

		if diffServiceMetadata(desiredService, svc) {
			desiredService.ResourceVersion = svc.ResourceVersion
			desiredService.Spec.ClusterIP = svc.Spec.ClusterIP

			if err := k.PutService(desiredService); err != nil {
				return err
			}
		}
	}

	return nil
}

func newService(service *apistructs.Service) *apiv1.Service {
	deployName := getDeployName(service)
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
				"app": deployName,
			},
		},
	}

	setServiceEnv(service, k8sService)

	for i, port := range service.Ports {
		k8sService.Spec.Ports = append(k8sService.Spec.Ports, apiv1.ServicePort{
			// TODO: name?
			Name: strutil.Concat(strings.ToLower(port.Protocol), "-", strconv.Itoa(i)),
			Port: int32(port.Port),
			// Dice上用户只填写Port, 即Port(service暴露的端口)和targetPort(容器暴露的端口)相同
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
func (k *Kubernetes) listServiceName(namespace string) ([]string, error) {
	strs := make([]string, 0)
	deployList, err := k.deploy.List(namespace)
	if err != nil {
		return strs, err
	}

	for _, item := range deployList.Items {
		strs = append(strs, item.Name)
	}

	daemonSets, err := k.ds.List(namespace)
	if err != nil {
		return strs, err
	}
	for _, item := range daemonSets.Items {
		strs = append(strs, item.Name)
	}

	return strs, nil
}

func getServiceName(service *apistructs.Service) string {
	if service.Env[ProjectNamespace] == "true" {
		return service.Env[ProjectNamespaceServiceNameNameKey]
	}
	return service.Name
}

func setServiceEnv(service *apistructs.Service, k8sService *apiv1.Service) {
	if v, ok := service.Env[ProjectNamespace]; ok && v == "true" {
		k8sService.Spec.Selector[LabelRuntimeID] = service.Env[KeyDiceRuntimeID]
	}
}

func (k *Kubernetes) deleteRuntimeServiceWithProjectNamespace(service apistructs.Service) error {
	if service.Env[ProjectNamespace] == "true" {
		err := k.service.Delete(service.Namespace, service.Env[ProjectNamespaceServiceNameNameKey])
		if err != nil {
			return fmt.Errorf("delete service %s error: %v", service.Name, err)
		}
	}
	if k.istioEngine != istioctl.EmptyEngine {
		err := k.istioEngine.OnServiceOperator(istioctl.ServiceDelete, &service)
		if err != nil {
			return fmt.Errorf("delete istio resource error: %v", err)
		}
	}
	return nil
}
