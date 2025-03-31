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

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/conf"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/istioctl"
)

// Two interfaces may call this function
// 1, Create
// 2, Update
func (k *Kubernetes) createOne(ctx context.Context, service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	if service == nil {
		return errors.Errorf("service empty")
	}
	// Step 1. Firstly create service
	// Only create k8s service for services with exposed ports
	if len(service.Ports) > 0 {
		if err := k.updateService(service, nil); err != nil {
			return err
		}
	}
	if service.ProjectServiceName != "" && len(service.Ports) > 0 {
		err := k.createProjectService(service, sg.ID)
		if err != nil {
			return err
		}
	}
	var err error
	switch service.WorkLoad {
	case types.ServicePerNode:
		err = k.createDaemonSet(ctx, service, sg)
	case types.ServiceJob:
		err = k.createJob(ctx, service, sg)
	default:
		// Step 2. Create related deployment
		err = k.createDeployment(ctx, service, sg)
	}
	if err != nil {
		return err
	}

	// TODO: Wait for the deployment running well ?
	// status, err := m.getDeployment(service)

	if k.istioEngine != istioctl.EmptyEngine {
		if err := k.istioEngine.OnServiceOperator(istioctl.ServiceCreate, service); err != nil {
			return err
		}
		if service.ProjectServiceName != "" {
			if projectService, ok := deepcopy.Copy(service).(*apistructs.Service); ok {
				projectService.Name = service.ProjectServiceName
				if err := k.istioEngine.OnServiceOperator(istioctl.ServiceCreate, projectService); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// not sure with whether this service exists
// bool The variable indicates whether it is really deleted
// Occurs in the following situations,
// 1, If the update interface analyzes the deletion, it is impossible to ensure the existence of k8s resources at this time
func (k *Kubernetes) tryDelete(namespace, name string) error {
	var (
		wg         sync.WaitGroup
		err1, err2 error
	)
	if k.istioEngine != istioctl.EmptyEngine {
		svc := &apistructs.Service{Namespace: namespace, Name: name}
		if err := k.istioEngine.OnServiceOperator(istioctl.ServiceDelete, svc); err != nil {
			return err
		}
	}
	wg.Add(2)
	go func() {
		err1 = k.deleteDeployment(namespace, name)
		wg.Done()
	}()
	go func() {
		err2 = k.deleteDaemonSet(namespace, name)
		wg.Done()
	}()
	wg.Wait()

	if err1 != nil && !util.IsNotFound(err1) {
		return errors.Errorf("failed to delete deployment, namespace: %s, name: %s, (%v)",
			namespace, name, err2)
	}

	if err2 != nil && !util.IsNotFound(err2) {
		return errors.Errorf("failed to delete daemonset, namespace: %s, name: %s, (%v)",
			namespace, name, err2)
	}

	return nil
}

// The creation operation needs to be completed before the update operation, because the newly created service may be a dependency of the service to be updated
// TODO: The updateOne function will be abstracted later
func (k *Kubernetes) updateOneByOne(ctx context.Context, sg *apistructs.ServiceGroup) error {
	labelSelector := make(map[string]string)
	var ns = sg.ProjectNamespace
	if ns == "" {
		ns = MakeNamespace(sg)
	} else {
		labelSelector[LabelServiceGroupID] = sg.ID
	}

	runtimeServiceMap, err := k.getRuntimeServiceMap(ns, labelSelector)
	if err != nil {
		//TODO:
		errMsg := fmt.Sprintf("list runtime service name err %v", err)
		logrus.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	if err := k.createImageSecretIfNotExist(ns); err != nil {
		return fmt.Errorf("create image secret when update one by one, err %v", err)
	}

	registryInfos := k.composeRegistryInfos(sg)
	if len(registryInfos) > 0 {
		err = k.UpdateImageSecret(ns, registryInfos)
		if err != nil {
			errMsg := fmt.Sprintf("failed to update secret %s on namespace %s, err: %v", conf.CustomRegCredSecret(), ns, err)
			logrus.Errorf(errMsg)
			return fmt.Errorf(errMsg)
		}
	}

	for _, svc := range sg.Services {
		svc.Namespace = ns
		runtimeServiceName := util.GetDeployName(&svc)
		// Existing in the old service collection, do the put operation
		// The visited record has been updated service
		if _, ok := runtimeServiceMap[runtimeServiceName]; ok {
			// firstly update the service
			// service is not the same as deployment, service is only created for services with exposed ports
			if err := k.updateService(&svc, nil); err != nil {
				return err
			}
			if err := k.updateProjectService(&svc, sg.ID); err != nil {
				return err
			}
			switch svc.WorkLoad {
			case types.ServicePerNode:
				desireDaemonSet, err := k.newDaemonSet(&svc, sg)
				if err != nil {
					return err
				}
				if err = k.updateDaemonSet(ctx, desireDaemonSet, &svc); err != nil {
					logrus.Debugf("failed to update daemonset in update interface, name: %s, (%v)", svc.Name, err)
					return err
				}
			case types.ServiceJob:
				err = k.createJob(ctx, &svc, sg)
			default:
				// then update the deployment
				desiredDeployment, err := k.newDeployment(&svc, sg)
				if err != nil {
					return err
				}
				if err = k.putDeployment(ctx, desiredDeployment, &svc); err != nil {
					logrus.Debugf("failed to update deployment in update interface, name: %s, (%v)", svc.Name, err)
					return err
				}
			}
			if k.istioEngine != istioctl.EmptyEngine {
				if err := k.istioEngine.OnServiceOperator(istioctl.ServiceUpdate, &svc); err != nil {
					return err
				}
			}
			runtimeServiceMap[runtimeServiceName] = RuntimeServiceRetain
		} else {
			// Does not exist in the old service collection, do the create operation
			// TODO: what to do if errors in Create ? before k8s create deployment ?
			// logrus.Debugf("in Update interface, going to create service(%s/%s)", ns, svc.Name)
			if err := k.createOne(ctx, &svc, sg); err != nil {
				logrus.Errorf("failed to create service in update interface, name: %s, (%v)", svc.Name, err)
				return err
			}
			runtimeServiceMap[runtimeServiceName] = RuntimeServiceRetain
		}
	}

	for svcName, operator := range runtimeServiceMap {
		if operator == RuntimeServiceDelete {
			if err := k.tryDelete(ns, svcName); err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to delete service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			svc, err := k.service.Get(ns, svcName)
			if err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to get k8s service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			if err = k.service.Delete(ns, svcName); err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to delete k8s service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			appLabelValue := svc.Spec.Selector["app"]
			deploys, err := k.deploy.List(ns, map[string]string{"app": appLabelValue})
			if err != nil {
				logrus.Errorf("failed to get deploys in ns %s", ns)
				return err
			}
			remainCount := 0
			for _, deploy := range deploys.Items {
				if deploy.DeletionTimestamp == nil {
					remainCount++
				}
			}
			if remainCount < 1 {
				err = k.service.Delete(ns, appLabelValue)
				if err != nil {
					if !util.IsNotFound(err) {
						logrus.Errorf("failed to delete global service %s in ns %s", svc.Name, ns)
						return err
					}
				}
			}
		}
	}
	return nil
}

func (k *Kubernetes) getStatelessStatus(ctx context.Context, sg *apistructs.ServiceGroup) (apistructs.StatusDesc, error) {
	var (
		failedReason string
		resultStatus apistructs.StatusDesc
		deploys      []appsv1.Deployment
		dsMap        map[string]appsv1.DaemonSet
		err          error
	)
	// init "unknown" status for each service
	for i := range sg.Services {
		sg.Services[i].Status = apistructs.StatusUnknown
		sg.Services[i].LastMessage = ""
	}

	var ns = MakeNamespace(sg)
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		k.setProjectServiceName(sg)
	}
	isReady := true
	isFailed := false

	if sg.ProjectNamespace != "" {
		deployList, err := k.deploy.List(ns, map[string]string{
			LabelServiceGroupID: sg.ID,
		})
		if err != nil {
			return apistructs.StatusDesc{}, fmt.Errorf("list deploy in ns %s err: %v", ns, err)
		}
		deploys = deployList.Items

		daemonsetExist := false

		for _, svc := range sg.Services {
			if svc.WorkLoad == types.ServicePerNode {
				daemonsetExist = true
				break
			}
		}
		if daemonsetExist {
			daemonsets, err := k.ds.List(ns, map[string]string{
				LabelServiceGroupID: sg.ID,
			})
			if err != nil {
				return apistructs.StatusDesc{}, fmt.Errorf("list daemonset in ns %s err: %v", ns, err)
			}

			dsMap = make(map[string]appsv1.DaemonSet, len(daemonsets.Items))
			for _, ds := range daemonsets.Items {
				dsMap[ds.Name] = ds
			}
		}

	} else {
		for _, svc := range sg.Services {
			deployList, err := k.deploy.List(ns, map[string]string{
				"app": svc.Name,
			})
			if err != nil {
				return apistructs.StatusDesc{}, fmt.Errorf("list deploy in ns %s err: %v", ns, err)
			}
			deploys = append(deploys, deployList.Items...)
		}
	}

	deployMap := make(map[string]appsv1.Deployment, len(deploys))
	for _, deploy := range deploys {
		deployMap[deploy.Name] = deploy
	}

	pods, err := k.pod.ListNamespacePods(ns)
	if err != nil {
		return apistructs.StatusDesc{}, err
	}

	for i := range sg.Services {
		var (
			status apistructs.StatusDesc
			err    error
		)
		switch sg.Services[i].WorkLoad {
		case types.ServicePerNode:
			status, err = k.getDaemonSetStatusFromMap(&sg.Services[i], dsMap)
		case types.ServiceJob:
			status, err = k.getJobStatusFromMap(&sg.Services[i], ns)

		default:
			// To distinguish the following exceptions：
			// 1, An error occurred during the creation process, and the entire runtime is deleted and then come back to query
			// 2, Others
			status, err = k.getDeploymentStatusFromMap(&sg.Services[i], deployMap)
		}
		if err != nil {
			// TODO: the state can be chanded to "Error"..
			status.Status = apistructs.StatusUnknown

			if !k8serror.NotFound(err) {
				return status, err
			}
			notfound, err := k.NotfoundNamespace(ns)
			if err != nil {
				errMsg := fmt.Sprintf("failed to get namespace existed info, namespace:%s, (%v)", ns, err)
				logrus.Errorf(errMsg)
				status.LastMessage = errMsg
				return status, err
			}

			// The namespace does not exist, indicating that there was an error during creation, and the runtime has been deleted by the scheduler
			if notfound {
				status.Status = apistructs.StatusErrorAndDeleted
				status.LastMessage = fmt.Sprintf("namespace not found, probably deleted, namespace: %s", ns)
			} else {
				// In theory, it will only appear in the process of deleting the namespace. A deployment has been deleted and the namespace is in terminating state and is about to be deleted.
				status.Status = apistructs.StatusUnknown
				status.LastMessage = fmt.Sprintf("found namespace exists but deployment not found,"+
					" namespace: %s, deployment: %s", sg.Services[i].Namespace, util.GetDeployName(&sg.Services[i]))
			}

			return status, err
		}
		sg.Services[i].ReadyReplicas = status.ReadyReplicas
		sg.Services[i].DesiredReplicas = status.DesiredReplicas
		if status.Status == apistructs.StatusFailed {
			isReady = false
			isFailed = true
			failedReason = status.Reason
			continue
		}
		if status.Status != apistructs.StatusReady {
			isReady = false
			resultStatus.Status = apistructs.StatusProgressing
			sg.Services[i].Status = apistructs.StatusProgressing
			podstatuses, err := k.pod.GetNamespacedPodsStatus(pods.Items, sg.Services[i].Name)
			if err != nil {
				logrus.Errorf("failed to get pod unready reasons, namespace: %v, name: %s, %v",
					sg.Services[i].Namespace,
					util.GetDeployName(&sg.Services[i]), err)
			}
			if len(podstatuses) != 0 {
				sg.Services[i].LastMessage = podstatuses[0].Message
				sg.Services[i].Reason = string(podstatuses[0].Reason)
			}
			continue
		}

		sg.Services[i].Status = apistructs.StatusHealthy
	}

	if isReady {
		resultStatus.Status = apistructs.StatusHealthy
	}
	if isFailed {
		resultStatus.Status = apistructs.StatusFailed
		resultStatus.LastMessage = failedReason
	}
	return resultStatus, nil
}

func (k *Kubernetes) getStatelessPodsStatus(sg *apistructs.ServiceGroup, svcName string) error {

	var ns = MakeNamespace(sg)
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		k.setProjectServiceName(sg)
	}

	pods, err := k.pod.ListNamespacePods(ns)
	if err != nil {
		return fmt.Errorf("list pods in ns %s err: %v", ns, err)
	}

	podRuntimeID := ""
	if runtimeID, ok := sg.Labels["GET_RUNTIME_STATELESS_SERVICE_POD_RUNTIME_ID"]; ok {
		podRuntimeID = runtimeID
	}
	serviceToPods := make(map[string][]apiv1.Pod)
	for i := range sg.Services {
		if sg.Services[i].Name != svcName {
			continue
		}

		for _, pod := range pods.Items {
			if pod.Labels == nil {
				continue
			}

			if !runtimeIDMatch(podRuntimeID, pod) {
				continue
			}

			serviceName := ""
			if _, ok := pod.Labels["DICE_SERVICE_NAME"]; !ok {
				serviceName = pod.Labels["DICE_SERVICE"]
			} else {
				serviceName = pod.Labels["DICE_SERVICE_NAME"]
			}

			if serviceName == "" {
				for _, v := range pod.Spec.Containers[0].Env {
					if v.Name == "DICE_SERVICE" || v.Name == "DICE_SERVICE_NAME" {
						serviceName = v.Value
						break
					}
				}
			}

			if serviceName == "" {
				continue
			}

			if serviceName == sg.Services[i].Name {
				if _, ok := serviceToPods[serviceName]; !ok {
					serviceToPods[serviceName] = make([]apiv1.Pod, 0)
				}
				serviceToPods[serviceName] = append(serviceToPods[serviceName], pod)
			}
		}

		if _, ok := serviceToPods[sg.Services[i].Name]; ok {
			if sg.Extra == nil {
				sg.Extra = make(map[string]string)
			}
			podsBytes, err := json.Marshal(serviceToPods[sg.Services[i].Name])
			if err != nil {
				return fmt.Errorf("json marshall service pods in ns %s for service %s err: %v", ns, sg.Services[i].Name, err)
			}

			sg.Extra[sg.Services[i].Name] = string(podsBytes)
		}

		break
	}

	return nil
}
