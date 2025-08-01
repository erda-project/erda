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
	"net/http"
	"os"
	"sync"
	"time"

	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/apistructs"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events/eventtypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/utils"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/wrapclient/edas"
	wrapclientset "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/wrapclient/kubernetes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/resourceinfo"
	executorutil "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	kind = "EDAS"
	// edas k8s service namespace
	defaultNamespace = metav1.NamespaceDefault
	notFound         = "not found"
)

// EDAS plugin's configure
func init() {
	_ = executortypes.Register(kind, func(name executortypes.Name, clustername string, options map[string]string, optionsPlus interface{}) (
		executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found edas address in env variables")
		}

		// key & secret of edas openAPI
		accessKey, ok := options["ACCESSKEY"]
		if !ok {
			return nil, errors.Errorf("not found edas accessKey in env variables")
		}
		accessSecret, ok := options["ACCESSSECRET"]
		if !ok {
			return nil, errors.Errorf("not found edas accessKey in env variables")
		}

		// EDAS Cluster ID
		clusterID, ok := options["CLUSTERID"]
		if !ok {
			return nil, errors.Errorf("not found edas clusterId in env variables")
		}

		// default: "cn-hangzhou"
		regionID, ok := options["REGIONID"]
		if !ok {
			regionID = "cn-hangzhou"
		}

		unlimitCPU, ok := options["UNLIMITCPU"]
		if !ok {
			unlimitCPU = "false"
		}

		// EDAS namespace, the default is the same as regionID
		logicalRegionID, ok := options["LOGICALREGIONID"]
		if !ok {
			logicalRegionID = ""
		}

		client, err := api.NewClientWithAccessKey(regionID, accessKey, accessSecret)
		if err != nil {
			return nil, errors.Wrap(err, "failed to new edas client with accessKey")
		}
		client.GetConfig().Transport = &http.Transport{
			DisableCompression: true,
		}

		k8sClient, err := k8sclient.New(clustername)
		if err != nil {
			return nil, errors.Errorf("get k8s client err %v", err)
		}

		regAddr, ok := options["REGADDR"]
		if !ok {
			return nil, errors.Errorf("not found dice registry addr in env variables")
		}

		l := logrus.WithField("executor", "edas")

		notifier, err := events.New(string(name), nil)
		if err != nil {
			l.Errorf("executor(%s) call eventbox new api error: %v", name, err)
			return nil, err
		}
		resourceInfo := resourceinfo.New(k8sClient.ClientSet)

		edas := &EDAS{
			l:              l,
			name:           name,
			options:        options,
			regAddr:        regAddr,
			notifier:       notifier,
			unlimitCPU:     unlimitCPU,
			resourceInfo:   resourceInfo,
			cs:             k8sClient.ClientSet,
			wrapEDASClient: edas.New(l, client, addr, clusterID, regionID, logicalRegionID, unlimitCPU),
			wrapClientSet:  wrapclientset.New(l, k8sClient.ClientSet, defaultNamespace),
		}

		if disableEvent := os.Getenv("DISABLE_EVENT"); disableEvent == "true" {
			return edas, nil
		}
		evCh := make(chan *eventtypes.StatusEvent, 10)
		// key is {runtimeNamespace}/{runtimeName}, value is spec.ServiceGroup
		lstore := &sync.Map{}
		stopCh := make(chan struct{}, 1)
		edas.registerEventChanAndLocalStore(evCh, stopCh, lstore)
		return edas, nil
	})
}

// EDAS edas server structure
type EDAS struct {
	l       *logrus.Entry
	name    executortypes.Name
	options map[string]string
	regAddr string

	notifier events.Notifier
	cs       kubernetes.Interface
	// Whether to limit the application of CPU resources less than 1c
	unlimitCPU     string
	resourceInfo   *resourceinfo.ResourceInfo
	wrapEDASClient edas.Interface
	wrapClientSet  wrapclientset.Interface
}

// Kind executor kind
func (e *EDAS) Kind() executortypes.Kind {
	return kind
}

// Name executor name
func (e *EDAS) Name() executortypes.Name {
	return e.name
}

// Create edas create runtime
func (e *EDAS) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	l := e.l.WithField("func", "Create")

	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("edas k8s create: invalid runtime spec")
	} else if err := utils.CheckRuntime(&runtime); err != nil {
		return nil, err
	}

	l.Debugf("create runtime, object: %+v", runtime)

	flows, err := executorutil.ParseServiceDependency(&runtime)
	if err != nil {
		return nil, errors.Wrapf(err, "parse service flow, runtime: %s",
			utils.CombineEDASAppGroup(runtime.Type, runtime.ID))
	}

	errChan := make(chan error, 1)
	go func() {
		if err = e.runAppFlow(context.Background(), flows, &runtime); err != nil {
			l.Errorf("failed to run runtime service flow: %v", err)

			// remove runtime & ignore error
			e.Remove(ctx, specObj)
			errChan <- err

			return
		}
	}()

	select {
	case err := <-errChan:
		close(errChan)
		return nil, err
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}

	return nil, nil
}

// Destroy edas destory runtime
func (e *EDAS) Destroy(ctx context.Context, specObj interface{}) error {
	return e.Remove(ctx, specObj)
}

// Status edas status of runtime
func (e *EDAS) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var (
		// Initialize it to prevent the upper console from being unrecognized
		status = apistructs.StatusDesc{
			Status: apistructs.StatusUnknown,
		}
		failReason  string
		lastMessage string
		isReady     = true
	)

	l := e.l.WithField("func", "Status")

	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return status, errors.New("edas k8s status: invalid runtime spec")
	}

	group := utils.CombineEDASAppGroup(runtime.Type, runtime.ID)
	for i, svc := range runtime.Services {
		appName := utils.CombineEDASAppNameWithGroup(group, svc.Name)

		svc.Namespace = runtime.Type
		// init status
		rtStatusDesc := apistructs.StatusDesc{
			Status: apistructs.StatusUnknown,
		}

		// check k8s deployment status
		deployStatus, err := e.getDeploymentStatus(appName)
		if err != nil {
			l.Errorf("get app %s deployment from k8s error: %+v", appName, err)

			runtime.Services[i].StatusDesc = rtStatusDesc
			isReady = false
			continue
		}

		// set replicas
		rtStatusDesc.ReadyReplicas = deployStatus.ReadyReplicas
		rtStatusDesc.DesiredReplicas = deployStatus.DesiredReplicas

		if deployStatus.Status != apistructs.StatusReady {
			l.Errorf("k8s deployment(%s) status is not ready: %+v", group, deployStatus.LastMessage)

			rtStatusDesc.LastMessage = fmt.Sprintf("deployment(%s) status is not ready, status: %s",
				runtime.ID, deployStatus.LastMessage)
			rtStatusDesc.Status = apistructs.StatusError
			rtStatusDesc.Reason = deployStatus.Reason
			runtime.Services[i].StatusDesc = rtStatusDesc

			isReady = false
			failReason = rtStatusDesc.Reason
			lastMessage = rtStatusDesc.LastMessage
			continue
		}

		if len(svc.Ports) != 0 {
			if _, err := e.wrapClientSet.GetK8sService(appName); err != nil {
				l.Errorf("k8s service status is not ready: %+v", status.LastMessage)

				rtStatusDesc.LastMessage = fmt.Sprintf("deployment(%s): service(%s) is not found or error", runtime.ID, appName)
				rtStatusDesc.Status = apistructs.StatusError
				runtime.Services[i].StatusDesc = rtStatusDesc

				isReady = false
				lastMessage = rtStatusDesc.LastMessage
				continue
			}
		}

		rtStatusDesc.Status = apistructs.StatusReady
		runtime.Services[i].StatusDesc = rtStatusDesc
	}

	//All services are ready, the runtime is set to ready, otherwise the state remains as the value during the calculation process
	if isReady {
		status.Status = apistructs.StatusReady
	} else {
		status.Status = apistructs.StatusError
		status.LastMessage = lastMessage
		status.Reason = failReason
	}

	return status, nil
}

// Remove edas remove runtime
// Because it takes a long time for edas to delete the interface, which exceeds the timeout set by the console, the app is deleted in parallel
func (e *EDAS) Remove(ctx context.Context, specObj interface{}) error {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return errors.New("edas k8s remove: invalid runtime spec")
	}

	group := utils.CombineEDASAppGroup(runtime.Type, runtime.ID)

	errChan := make(chan error, 1)
	go func(ss []apistructs.Service) {
		for _, srv := range ss {
			// TODO: how to handle the error
			if err := e.removeService(ctx, group, &srv); err != nil {
				errChan <- err
			}
		}
		close(errChan)
	}(runtime.Services)

	// HACK: edas api Inevitable timeout (> 10s), here you need to wait for 5s
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}

	return nil
}

// Update edas update runtime
func (e *EDAS) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	l := e.l.WithField("func", "Update")

	var oldRun apistructs.ServiceGroup

	newRun, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("edas k8s update: invalid runtime spec")
	} else if err := utils.CheckRuntime(&newRun); err != nil {
		return nil, err
	}

	// Get the deploy list from k8s API
	l.Debugf("update runtime, new runtime: %+v", newRun)

	group := utils.CombineEDASAppGroup(newRun.Type, newRun.ID)
	if err := e.wrapClientSet.GetK8sDeployList(group, &oldRun.Services); err != nil {
		l.Debugf("get deploy from k8s error: %+v", err)
		return nil, err
	}

	l.Debugf("old runtime service list is : %+v", oldRun.Services)
	if err := e.cyclicUpdateService(ctx, &newRun, &oldRun); err != nil {
		l.Debugf("cyclic update service error: %+v", err)
		return nil, err
	}

	return nil, nil
}

// Inspect Query runtime information
func (e *EDAS) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	l := e.l.WithField("func", "Inspect")

	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("edas k8s inspect: invalid runtime spec")
	}

	if serviceName, ok := runtime.Labels["GET_RUNTIME_STATELESS_SERVICE_POD"]; ok && serviceName != "" {
		pods, err := e.cs.CoreV1().Pods(defaultNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", utils.CombineEDASAppName(runtime.Type, runtime.ID, serviceName)),
		})
		if err != nil {
			return nil, fmt.Errorf("get pods for servicegroup %+v failed: %v", runtime, err)
		}

		if runtime.Extra == nil {
			runtime.Extra = make(map[string]string)
		}

		podsBytes, err := json.Marshal(pods.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to json marshall edas service pods in ns %s for service %s err: %v",
				defaultNamespace, serviceName, err)
		}
		runtime.Extra[serviceName] = string(podsBytes)
		return &runtime, nil
	}

	// Metadata information is passed in from the upper layer, here you only need to get the state of the runtime and assemble it into the runtime to return
	status, err := e.Status(ctx, specObj)
	if err != nil {
		return nil, errors.Errorf("edas k8s inspect: %v", err)
	}

	l.Infof("Inspect runtime(%s) status: %+v", runtime.ID, status)

	group := utils.CombineEDASAppGroup(runtime.Type, runtime.ID)
	runtime.Status = status.Status
	runtime.LastMessage = status.LastMessage

	for i, svc := range runtime.Services {
		appName := utils.CombineEDASAppNameWithGroup(group, svc.Name)

		if len(svc.Ports) > 0 {
			kubeSvc, err := e.wrapClientSet.GetK8sService(appName)
			if err != nil {
				l.Warnf("failed to inspect runtime(%s), service name: %s, error: %v",
					group, appName, err)
			} else {
				svcRecord := kubeSvc.Name + ".default.svc.cluster.local"
				runtime.Services[i].ProxyIp = svcRecord
				runtime.Services[i].ProxyPorts = diceyml.ComposeIntPortsFromServicePorts(svc.Ports)
				runtime.Services[i].Vip = svcRecord
			}
		}
	}

	return &runtime, nil
}

func (e *EDAS) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, nil
}
func (e *EDAS) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}
func (e *EDAS) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in EDAS")
}

func (e *EDAS) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

func (e *EDAS) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{
		CPUOverCommit:        1.0,
		ProdCPUOverCommit:    1.0,
		DevCPUOverCommit:     1.0,
		TestCPUOverCommit:    1.0,
		StagingCPUOverCommit: 1.0,
		ProdMEMOverCommit:    1.0,
		DevMEMOverCommit:     1.0,
		TestMEMOverCommit:    1.0,
		StagingMEMOverCommit: 1.0,
		Nodes: map[string]*apistructs.NodeResourceInfo{
			"1.1.1.1": {
				Labels:         []string{},
				IgnoreLabels:   true,
				Ready:          true,
				CPUAllocatable: 999999999999999.0,
				MemAllocatable: 999999999999999,
			},
		},
	}, nil
}

func (e *EDAS) CleanUpBeforeDelete() {}
func (e *EDAS) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for edas")
}
func (e *EDAS) KillPod(podname string) error {
	return fmt.Errorf("not support for edas")
}

func (e *EDAS) Scale(ctx context.Context, specObj interface{}) (interface{}, error) {
	l := e.l.WithField("func", "Scale")

	sg, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		errMsg := fmt.Sprintf("edas k8s scale: invalid service group spec")
		l.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	if _, ok = sg.Labels[pstypes.ErdaPALabelKey]; ok {
		errMsg := fmt.Sprintf("edas k8s scale: not support sg with label %s", pstypes.ErdaPALabelKey)
		l.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	// only support scale one service resources
	if len(sg.Services) != 1 {
		errMsg := fmt.Sprintf("the scaling service count is not equal 1")
		l.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if err := utils.CheckRuntime(&sg); err != nil {
		errMsg := fmt.Sprintf("check the runtime struct err: %v", err)
		l.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	services := []apistructs.Service{}
	destService := sg.Services[0]
	originService := &apistructs.Service{}

	appName := utils.CombineEDASAppName(sg.Type, sg.ID, destService.Name)
	var (
		appID string
		err   error
	)
	if appID, err = e.wrapEDASClient.GetAppID(appName); err != nil {
		errMsg := fmt.Sprintf("get appID err in scale: %v", err)
		l.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if e.unlimitCPU == "true" {
		if destService.Resources.Cpu < 1 {
			destService.Resources.Cpu = 0
		} else {
			destService.Resources.Cpu = math.Floor(destService.Resources.Cpu + 0.5)
		}
	}

	group := utils.CombineEDASAppGroup(sg.Type, sg.ID)

	l.Infof("start to get k8s deploy list %s", group)
	if err = e.wrapClientSet.GetK8sDeployList(group, &services); err != nil {
		l.Debugf("get deploy from k8s error: %+v", err)
		return nil, err
	}

	for _, service := range services {
		if service.Name == destService.Name {
			service.Resources.Cpu = service.Resources.Cpu / 1000
			service.Resources.Mem = service.Resources.Mem / 1024 / 1024
			originService = &service
			break
		}
	}

	var errString = make(chan string, 0)
	go func() {
		if originService != nil {
			// Query the latest release order, and terminate if it is running
			orderList, _ := e.wrapEDASClient.ListRecentChangeOrderInfo(appID)
			if len(orderList.ChangeOrder) > 0 && orderList.ChangeOrder[0].Status == 1 {
				err = e.wrapEDASClient.AbortChangeOrder(orderList.ChangeOrder[0].ChangeOrderId)
				if err != nil {
					errMsg := fmt.Sprintf("scale k8s application err: %v", err)
					l.Errorf(errMsg)
					errString <- errMsg
				}
			}

			l.Infof("diff cpu %v, origin is %v, dest is %v", originService.Resources.Cpu == destService.Resources.Cpu, originService.Resources.Cpu, destService.Resources.Cpu)
			l.Infof("diff memory %v, origin is %v, dest is %v", originService.Resources.Mem == destService.Resources.Mem, originService.Resources.Mem, destService.Resources.Mem)
			l.Infof("diff scale %v, origin is %v, dest is %v", originService.Scale == destService.Scale, originService.Scale, destService.Scale)
			if originService.Resources.Cpu == destService.Resources.Cpu &&
				originService.Resources.Mem == destService.Resources.Mem &&
				originService.Scale != destService.Scale {

				if err := e.wrapEDASClient.ScaleApp(appID, destService.Scale); err != nil {
					l.Error(err)
					errString <- err.Error()
				}
			} else {
				spec, err := e.fillServiceSpec(&destService, &sg, true)
				if err != nil {
					errMsg := fmt.Sprintf("compose service Spec application err: %v", err)
					l.Errorf(errMsg)
					errString <- errMsg
				}
				err = e.wrapEDASClient.DeployApp(appID, spec)
				if err != nil {
					errMsg := fmt.Sprintf("compose service Spec application err: %v", err)
					l.Errorf(errMsg)
					errString <- errMsg
				}
			}
			errString <- ""
		}
		errString <- fmt.Sprintf("not found service %s", destService.Name)
	}()
	select {
	case str := <-errString:
		if str != "" {
			return nil, fmt.Errorf(str)
		}
	case <-time.After(5 * time.Second):
	}
	return sg, nil
}
