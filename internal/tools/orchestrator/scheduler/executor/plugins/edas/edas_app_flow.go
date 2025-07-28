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

package edas

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/utils"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
	executorutil "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *EDAS) runAppFlow(ctx context.Context, flows [][]*apistructs.Service, runtime *apistructs.ServiceGroup) error {
	l := e.l.WithField("func", "runAppFlow")

	group := utils.CombineEDASAppGroup(runtime.Type, runtime.ID)

	for i, batch := range flows {
		l.Infof("create runtime: %s run batch %d %+v", group, i+1, batch)

		for _, s := range batch {
			go func() {
				var service *apistructs.Service

				service = s
				l.Infof("run app flow to create service %s", s.Name)
				if err := e.createService(ctx, runtime, service); err != nil {
					l.Errorf("failed to create service: %s, error: %v",
						utils.CombineEDASAppNameWithGroup(group, s.Name), err)
				}
			}()
			time.Sleep(1 * time.Second)
		}

		if err := e.waitRuntimeRunningOnBatch(ctx, batch, group); err != nil {
			return errors.Wrap(err, "wait service flow on batch")
		}
	}
	l.Infof("run app flow %s finished", group)
	return nil
}

func (e *EDAS) createService(ctx context.Context, sg *apistructs.ServiceGroup, s *apistructs.Service) error {
	l := e.l.WithField("func", "createService")

	serviceSpec, err := e.fillServiceSpec(s, sg, false)
	if err != nil {
		return errors.Wrap(err, "fill service spec")
	}

	// Create application
	if _, err = e.wrapEDASClient.InsertK8sApp(serviceSpec); err != nil {
		return errors.Wrap(err, "edas create app")
	}

	appName := utils.CombineEDASAppName(sg.Type, sg.ID, s.Name)

	//create k8s service
	if err = e.wrapClientSet.CreateK8sService(appName, sg.ID, s.Name, diceyml.ComposeIntPortsFromServicePorts(s.Ports)); err != nil {
		l.Errorf("failed to create k8s service, appName: %s, error: %v", appName, err)
		return errors.Wrap(err, "edas create k8s service")
	}

	return nil
}

func (e *EDAS) updateService(ctx context.Context, sg *apistructs.ServiceGroup, s *apistructs.Service) error {
	l := e.l.WithField("func", "updateService")

	appName := utils.CombineEDASAppName(sg.Type, sg.ID, s.Name)

	// Check whether the service exists, if it does not exist, create a new one; otherwise, update it
	if appID, err := e.wrapEDASClient.GetAppID(appName); err != nil {
		if err.Error() == notFound {
			l.Warningf("app(%s) is not found via edas api, will create it ", appName)
			if err = e.createService(ctx, sg, s); err != nil {
				l.Errorf("failed to create service(%s): %v", appName, err)
				return err
			}
		} else {
			l.Errorf("failed to query app(%s) by update service: %s", appName, err)
			return err
		}
	} else {
		//查询最新一次的发布单，如果存在运行中则终止
		orderList, _ := e.wrapEDASClient.ListRecentChangeOrderInfo(appID)
		if len(orderList.ChangeOrder) > 0 && orderList.ChangeOrder[0].Status == 1 {
			e.wrapEDASClient.AbortChangeOrder(orderList.ChangeOrder[0].ChangeOrderId)
		}

		svcSpec, err := e.fillServiceSpec(s, sg, true)
		if err != nil {
			return errors.Wrap(err, "fill service spec")
		}

		if err = e.wrapEDASClient.DeployApp(appID, svcSpec); err != nil {
			l.Errorf("failed to deploy app: %s, error: %v", appName, err)
			return err
		}

		if err := e.wrapClientSet.CreateOrUpdateK8sService(ctx, appName, sg.ID, s.Name, diceyml.ComposeIntPortsFromServicePorts(s.Ports)); err != nil {
			l.Errorf("failed to update k8s service, appName: %s, error: %v", appName, err)
			return errors.Wrap(err, "edas update k8s service")
		}
	}

	return nil
}

// TODO: how to handle the error
func (e *EDAS) removeService(ctx context.Context, group string, s *apistructs.Service) error {
	l := e.l.WithField("func", "removeService")

	appName := utils.CombineEDASAppGroup(group, s.Name)
	if err := e.wrapEDASClient.DeleteAppByName(appName); err != nil {
		l.Errorf("failed to delete app(%s): %v", appName, err)
		return err
	}

	if err := e.wrapClientSet.DeleteK8sService(appName); err != nil {
		l.Errorf("failed to delete k8s svc of app(%s): %v", appName, err)
		return err
	}
	// HACK: Regardless of whether calling edas api to delete the service is successful, try to delete the related service directly through k8s
	// if err = e.deleteDeploymentAndService(group, s); err != nil {
	// 	l.Warnf("failed to delete k8s deployments and service, appName: %s, error: %v", appName, err)
	// }

	return nil
}

func (e *EDAS) cyclicUpdateService(ctx context.Context, newRuntime, oldRuntime *apistructs.ServiceGroup) error {
	l := e.l.WithField("func", "cyclicUpdateService")

	errChan := make(chan error, 1)
	group := utils.CombineEDASAppGroup(newRuntime.Type, newRuntime.ID)

	// Resolve dependencies
	flows, err := executorutil.ParseServiceDependency(newRuntime)
	if err != nil {
		return errors.Wrapf(err, "parse service flow, runtime: %s", group)
	}
	go func() {
		l.Debugf("start to cyclicUpdateService, group: %s", group)
		defer l.Debugf("end cyclicUpdateService, group: %s", group)

		// Detect services that need to be deleted in advance
		// 1. The service whose name has been deleted or updated
		// 2. The service whose port has been modified
		svcs := checkoutServicesToDelete(newRuntime, oldRuntime)
		for _, svc := range *svcs {
			appName := utils.CombineEDASAppNameWithGroup(group, svc.Name)
			l.Warningf("need to delete service(%s) because the user modified name or ports !!!", appName)

			if err := e.removeService(ctx, group, &svc); err != nil {
				l.Errorf("failed to remove service by cyclic update: %s, error: %v", appName, err)
			}
		}

		for _, batch := range flows {
			for _, newSvc := range batch {
				var ok bool
				var oldSvc *apistructs.Service

				svcName := newSvc.Name
				appName := utils.CombineEDASAppNameWithGroup(group, svcName)
				// add service
				if ok, oldSvc = isServiceInRuntime(svcName, oldRuntime); !ok || oldSvc == nil {
					l.Infof("cyclicupdate to create service %s", svcName)
					if err = e.createService(ctx, newRuntime, newSvc); err != nil {
						l.Errorf("failed to create service: %s, error: %v", appName, err)
						errChan <- err
						return
					}
					continue
				}

				// update service
				// Does not include domain name updates
				if err = e.updateService(ctx, newRuntime, newSvc); err != nil {
					l.Errorf("failed to update service: %s, error: %v", appName, err)
					errChan <- err
					return
				}
			}
		}
	}()

	// HACK: The edas api must time out (> 10s). The reason for waiting for 5s here is to facilitate the status update of the runtime.
	// Prevent the upper layer from querying the status when the asynchronous has not been executed.
	select {
	case err := <-errChan:
		close(errChan)
		return err
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}

	return nil
}

// FIXME: Is 20 times reasonable?
func (e *EDAS) waitRuntimeRunningOnBatch(ctx context.Context, batch []*apistructs.Service, group string) error {
	l := e.l.WithField("func", "waitRuntimeRunningOnBatch")

	for i := 0; i < 60; i++ {
		done := map[string]struct{}{}

		time.Sleep(10 * time.Second)
		for _, srv := range batch {
			if _, ok := done[srv.Name]; ok {
				continue
			}
			appName := utils.CombineEDASAppNameWithGroup(group, srv.Name)
			// 1. Confirm app status from edas
			status, err := e.wrapEDASClient.QueryAppStatus(appName)
			if err != nil {
				l.Errorf("failed to query app(name: %s) status: %v", appName, err)
				continue
			}
			if status != types.AppStatusFailed {
				// 2. After app status is equal to running, confirm whether the k8s service is ready
				if len(srv.Ports) == 0 {
					done[appName] = struct{}{}
					continue
				}

				if _, err = e.wrapClientSet.GetK8sService(appName); err != nil {
					l.Errorf("failed to get k8s service, appName: %s, error: %v", appName, err)
					continue
				}

				done[appName] = struct{}{}
			}
		}

		if len(done) == len(batch) {
			l.Infof("successfully to wait runtime running on batch")
			return nil
		}
	}

	return errors.Errorf("failed to wait runtime(%s) status to running on batch.", group)
}

// Confirm whether to delete the service list
// The conditions are as follows:
// 1.Deleted service
// 2.The service whose name has been changed
// 3.The service of the modified port
func checkoutServicesToDelete(newRuntime, oldRuntime *apistructs.ServiceGroup) *[]apistructs.Service {
	var svcs []apistructs.Service

	if newRuntime == nil || oldRuntime == nil {
		return nil
	}

	for _, oldSvc := range oldRuntime.Services {
		ok, newSvc := isServiceInRuntime(oldSvc.Name, newRuntime)
		if !ok || (oldSvc.Ports[0].Port != newSvc.Ports[0].Port) {
			svcs = append(svcs, oldSvc)
		}
	}

	return &svcs
}

func isServiceInRuntime(name string, run *apistructs.ServiceGroup) (bool, *apistructs.Service) {
	if name == "" || run == nil {
		logrus.Warningf("hasServiceInRuntime invalid params, name or runtime is null")
		return false, nil
	}

	for _, svc := range run.Services {
		if svc.Name == name {
			return true, &svc
		}
	}

	return false, nil
}

func (e *EDAS) generateServiceEnvs(s *apistructs.Service, runtime *apistructs.ServiceGroup) (map[string]string, error) {
	var envs map[string]string

	group, appName := utils.CombineEDASAppInfo(runtime.Type, runtime.ID, s.Name)
	envs = s.Env
	if envs == nil {
		envs = make(map[string]string, 10)
	}

	addEnv := func(s *apistructs.Service, envs *map[string]string) error {
		appName := utils.CombineEDASAppNameWithGroup(group, s.Name)
		kubeSvc, err := e.wrapClientSet.GetK8sService(appName)
		if err != nil {
			return err
		}

		svcRecord := kubeSvc.Name + ".default.svc.cluster.local"
		// add {serviceName}_HOST
		key := utils.MakeEnvVariableName(s.Name) + "_HOST"
		(*envs)[key] = svcRecord

		// {serviceName}_PORT Point to the first port
		if len(s.Ports) > 0 {
			key := utils.MakeEnvVariableName(s.Name) + "_PORT"
			(*envs)[key] = strconv.Itoa(s.Ports[0].Port)
		}

		//If there are multiple ports, use them in sequence, like {serviceName}_PORT0, {serviceName}_PORT1,...
		for i, port := range s.Ports {
			key := utils.MakeEnvVariableName(s.Name) + "_PORT" + strconv.Itoa(i)
			(*envs)[key] = strconv.Itoa(port.Port)
		}

		return nil
	}

	// TODO: All containers have all service environment variables for addon deployment
	// Since edas needs to create the app before it can determine the service name, it is impossible to pre-fill the full amount of environment variables
	// edas does not deploy addons and does not support GLOBAL; however, it is necessary to deploy low-profile addons for testing, so no error is returned.
	if runtime.ServiceDiscoveryMode == "GLOBAL" {
		//return nil, errors.Errorf("not support ServiceDiscoveryMode: %v", runtime.ServiceDiscoveryMode)
	} else {
		for _, name := range s.Depends {
			var depSvc *apistructs.Service
			for _, svc := range runtime.Services {
				if svc.Name == name {
					depSvc = &svc
					break
				}
			}
			// not found
			if depSvc == nil {
				return nil, errors.Errorf("not find service: %s", name)
			}

			if len(depSvc.Ports) == 0 {
				continue
			}

			//Inject {depSvc}_HOST, {depSvc}_PORT, etc.
			if err := addEnv(depSvc, &envs); err != nil {
				return nil, err
			}

			depAppName := utils.CombineEDASAppNameWithGroup(group, depSvc.Name)
			if s.Labels["IS_ENDPOINT"] == "true" && len(depSvc.Ports) > 0 {
				kubeSvc, err := e.wrapClientSet.GetK8sService(depAppName)
				if err != nil {
					return nil, err
				}
				svcRecord := kubeSvc.Name + ".default.svc.cluster.local"
				port := depSvc.Ports[0]
				envs["BACKEND_URL"] = svcRecord + ":" + strconv.Itoa(port.Port)
			}
		}
	}

	svcAddr := appName + ".default.svc.cluster.local"
	// add K8S label
	envs["IS_K8S"] = "true"
	// add svc label
	envs["SELF_HOST"] = svcAddr
	if len(s.Ports) != 0 {
		envs["SELF_PORT"] = strconv.Itoa(s.Ports[0].Port)
		envs["SELF_URL"] = "http://" + svcAddr + ":" + strconv.Itoa(s.Ports[0].Port)
		envs["SELF_PORT0"] = strconv.Itoa(s.Ports[0].Port)
	}

	// TODO: add self env
	//Problem: After the service is created, there will be k8s service, which makes it impossible to insert SELF_HOST env in advance

	return envs, nil
}

func (e *EDAS) fillServiceSpec(s *apistructs.Service, sg *apistructs.ServiceGroup, isUpdate bool) (*types.ServiceSpec, error) {
	var envs map[string]string
	var err error

	l := e.l.WithField("func", "fillServiceSpec")

	appName := utils.CombineEDASAppName(sg.Type, sg.ID, s.Name)

	l.Debugf("start to fill service spec: %s", appName)

	svcSpec := &types.ServiceSpec{
		Name:      appName,
		Instances: s.Scale,
		Mem:       int(s.Resources.Mem),
		Ports:     diceyml.ComposeIntPortsFromServicePorts(s.Ports),
	}

	// FIXME: hacking for registry
	// e.g. registry.marathon.l4lb.thisdcos.directory:5000  docker-registry.registry.marathon.mesos:5000
	body := strings.Split(s.Image, ":5000")
	if len(body) > 1 && e.regAddr != "" {
		svcSpec.Image = e.regAddr + body[1]
	} else {
		svcSpec.Image = s.Image
	}

	//For clusters with unlimitCPU turned on, the processing of less than 1 core is unlimited, and the default limit is 1C
	if e.unlimitCPU == "true" {
		svcSpec.CPU = 0
	} else {
		svcSpec.CPU = 1
	}

	//edas mount nfs volume
	if len(s.Binds) != 0 {
		type localVolumes struct {
			Type      string `json:"type"`
			NodePath  string `json:"nodePath"`
			MountPath string `json:"mountPath"`
		}
		var lv []localVolumes
		var lvBody []byte
		for _, bind := range s.Binds {
			if bind.HostPath == "" || bind.ContainerPath == "" || !strutil.HasPrefixes(bind.HostPath, "/") {
				continue
			}
			lv = append(lv, localVolumes{
				Type:      "DirectoryOrCreate",
				NodePath:  bind.HostPath,
				MountPath: bind.ContainerPath,
			})
		}
		lvBody, err = json.Marshal(lv)
		if err != nil {
			return nil, errors.Wrapf(err, "json marshal service localvolume")
		}
		l.Debugf("fill service spec localvolume args: %s", string(lvBody))
		svcSpec.LocalVolume = string(lvBody)
	}
	if s.Resources.Cpu >= 1.0 {
		cpu := math.Floor(s.Resources.Cpu + 0.5)
		svcSpec.CPU = int(cpu)
	}
	svcSpec.Mcpu = int(s.Resources.Cpu * 1000)
	// command: sh
	// args: [{"argument":"-c"},{"argument":"test"}]
	if len(s.Cmd) != 0 {
		svcSpec.Cmd = "sh"
		// inputArgs := strings.Split("-c "+s.Cmd, " ")
		inputArgs := []string{"-c", s.Cmd}

		type cArg struct {
			Argument string `json:"argument"`
		}
		var cArgs []cArg
		var argBody []byte
		for _, inputArg := range inputArgs {
			cArgs = append(cArgs, cArg{Argument: inputArg})
		}
		if isUpdate {
			argBody, err = json.Marshal(inputArgs)
		} else {
			argBody, err = json.Marshal(cArgs)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "json marshal service cmd")
		}

		l.Debugf("fill service spec command args: %s", string(argBody))
		svcSpec.Args = string(argBody)
	}

	if envs, err = e.generateServiceEnvs(s, sg); err != nil {
		l.Errorf("failed to generate service envs: %s, error: %s", appName, err)
		return nil, err
	}

	// envs: [{"name":"testkey","value":"testValue"}]
	if len(envs) != 0 {
		envString, err := utils.EnvToString(envs)
		if err != nil {
			return nil, err
		}
		svcSpec.Envs = envString
	}

	// annotations: {"annotation-name-1":"annotation-value-1","annotation-name-2":"annotation-value-2"}
	if err = setAnnotations(svcSpec, envs); err != nil {
		return nil, errors.Wrapf(err, "failed to set annotations, service name: %s", appName)
	}

	if err = setLabels(svcSpec, sg.ID, s.Name); err != nil {
		return nil, errors.Wrapf(err, "failed to set labels, service name: %s", appName)
	}

	if err = setHealthCheck(s, svcSpec); err != nil {
		return nil, errors.Wrapf(err, "failed to set health check, service name: %s", appName)
	}

	// TODO: support postStart, preStop, nasId, mountDescs, storageType, localvolume

	l.Infof("fill service spec: %+v", svcSpec)

	return svcSpec, nil
}

func (e *EDAS) getDeploymentStatus(appName string) (apistructs.StatusDesc, error) {
	status := apistructs.StatusDesc{
		Status: apistructs.StatusUnknown,
	}

	dep, err := e.wrapEDASClient.GetAppDeployment(appName)
	if err != nil {
		return status, err
	}

	dps := dep.Status
	status.DesiredReplicas = pointer.Int32Deref(dep.Spec.Replicas, 0)
	status.ReadyReplicas = dps.ReadyReplicas
	// this would not happen in theory
	if len(dps.Conditions) == 0 {
		status.Status = apistructs.StatusUnknown
		status.LastMessage = "could not get status condition"
		return status, nil
	}

	for _, condition := range dps.Conditions {
		if condition.Type == k8sapi.DeploymentReplicaFailure && condition.Status == "True" {
			status.Status = apistructs.StatusFailing
			status.LastMessage = condition.Message
			status.Reason = condition.Reason
			return status, nil
		}
		if condition.Type == k8sapi.DeploymentAvailable && condition.Status == "False" {
			status.Status = apistructs.StatusFailing
			status.LastMessage = condition.Message
			status.Reason = condition.Reason
			return status, nil
		}
	}

	if dps.Replicas == dps.ReadyReplicas &&
		dps.Replicas == dps.AvailableReplicas &&
		dps.Replicas == dps.UpdatedReplicas {
		if dps.Replicas > 0 {
			status.Status = apistructs.StatusReady
			status.LastMessage = fmt.Sprintf("deployment(%s) is running", dep.Name)
		} else {
			// This state is only present at the moment of deletion
			status.LastMessage = fmt.Sprintf("deployment(%s) replica is 0", dep.Name)
		}
	}

	return status, nil
}

func setHealthCheck(service *apistructs.Service, svcSpec *types.ServiceSpec) error {
	var (
		b   []byte
		err error
	)

	probe := k8s.FillHealthCheckProbe(service)
	if probe != nil {
		if b, err = json.Marshal(probe); err != nil {
			return err
		}

		svcSpec.Liveness = string(b)
		svcSpec.Readiness = string(b)
	}

	return nil
}

func setAnnotations(svcSpec *types.ServiceSpec, envs map[string]string) error {
	if svcSpec == nil {
		return errors.New("service spec is nil")
	}

	annotations := make(map[string]string)
	for k, v := range envs {
		if annotationKey := executorutil.ParseAnnotationFromEnv(k); annotationKey != "" {
			annotations[annotationKey] = v
		}
	}

	if len(annotations) == 0 {
		return nil
	}

	ret, err := json.Marshal(annotations)
	if err != nil {
		return err
	}
	svcSpec.Annotations = string(ret)
	return nil
}

func setLabels(svcSpec *types.ServiceSpec, sgID, serviceName string) error {
	if svcSpec == nil {
		return errors.New("service spec is nil")
	}

	targetLabels := map[string]string{
		"app":             serviceName,
		"servicegroup-id": sgID,
	}

	ret, err := json.Marshal(targetLabels)
	if err != nil {
		return err
	}
	svcSpec.Labels = string(ret)
	return nil
}
