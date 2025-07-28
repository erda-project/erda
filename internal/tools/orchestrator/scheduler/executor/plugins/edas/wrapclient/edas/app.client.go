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
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/utils"
)

// k8s min ready seconds
var minReadySeconds = 30

// GetAppID get application by id
func (c *wrapEDAS) GetAppID(appName string) (string, error) {
	l := c.l.WithField("func", "GetAppID")

	req := api.CreateListApplicationRequest()
	req.SetDomain(c.addr)
	req.AppName = appName

	// get application list
	resp, err := c.client.ListApplication(req)
	if err != nil {
		return "", errors.Wrap(err, "list application")
	}

	l.Info("request id: ", resp.RequestId)

	for _, app := range resp.ApplicationList.Application {
		if app.Name == appName {
			l.Infof("successfully to get app id: %s, name: %s", app.AppId, appName)
			return app.AppId, nil
		}
	}

	return "", ErrApplicationNotFound
}

// deleteAppByID delete application by app id
func (c *wrapEDAS) deleteAppByID(id string) error {
	l := c.l.WithField("func", "deleteAppByID")

	l.Infof("start to delete app by id: %s", id)

	// stop application first
	if err := c.stopAppByID(id); err != nil {
		return err
	}

	req := api.CreateDeleteK8sApplicationRequest()
	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.SetDomain(c.addr)
	req.AppId = id

	// DeleteApplicationRequest
	resp, err := c.client.DeleteK8sApplication(req)
	if err != nil {
		return errors.Errorf("response http context: %s, error: %v", resp.GetHttpContentString(), err)
	}

	l.Info("request id: ", resp.RequestId)
	l.Debugf("delete app(%s) response, requestID: %s, code: %d, message: %s, changeOrderID: %s",
		id, resp.RequestId, resp.Code, resp.Message, resp.ChangeOrderId)

	if len(resp.ChangeOrderId) != 0 {
		status, err := c.LoopTerminationStatus(resp.ChangeOrderId)
		if err != nil {
			return errors.Wrapf(err, "get delete status by loop")
		}

		l.Infof("loop termination status from delete application, order id: %s, status: %s",
			resp.ChangeOrderId, types.ChangeOrderStatusString[status])

		if status != types.CHANGE_ORDER_STATUS_SUCC {
			return errors.Errorf("failed to get the status of deleting app(%s), status = %s", id, types.ChangeOrderStatusString[status])
		}
	}

	l.Infof("successfully to delete app by id: %s", id)
	return nil
}

// stopAppByID stop application by id
func (c *wrapEDAS) stopAppByID(id string) error {
	l := c.l.WithField("func", "stopAppByID")

	l.Infof("stop app first, id: %s", id)

	stopReq := api.CreateStopK8sApplicationRequest()
	stopReq.Headers = utils.AppendCommonHeaders(stopReq.Headers)
	stopReq.SetDomain(c.addr)
	stopReq.AppId = id

	stopResp, err := c.client.StopK8sApplication(stopReq)
	if err != nil {
		return errors.Errorf("response http context: %s, error: %v", stopResp.GetHttpContentString(), err)
	}

	l.Info("request id: ", stopResp.RequestId)

	l.Debugf("stop app(%s) response, requestID: %s, code: %d, message: %s, changeOrderID: %s",
		id, stopResp.RequestId, stopResp.Code, stopResp.Message, stopResp.ChangeOrderId)

	if len(stopResp.ChangeOrderId) != 0 {
		l.Infof("start to load stop change order status, change order id: %s", stopResp.ChangeOrderId)

		status, err := c.LoopTerminationStatus(stopResp.ChangeOrderId)
		if err != nil {
			return errors.Wrapf(err, "get stop status by loop")
		}

		l.Infof("loop termination status from stop application, order id: %s, status: %s",
			stopResp.ChangeOrderId, types.ChangeOrderStatusString[status])

		if status != types.CHANGE_ORDER_STATUS_SUCC {
			return errors.Errorf("failed to get the status of stopping app(%s), status = %s",
				id, types.ChangeOrderStatusString[status])
		}
	}

	l.Infof("successfully to stop app by id: %s", id)
	return nil
}

// QueryAppStatus query application status
func (c *wrapEDAS) QueryAppStatus(appName string) (types.AppStatus, error) {
	l := c.l.WithField("func", "QueryAppStatus")

	var state = types.AppStatusRunning

	appID, err := c.GetAppID(appName)
	if err != nil {
		return state, err
	}

	req := api.CreateQueryApplicationStatusRequest()
	req.SetDomain(c.addr)
	req.AppId = appID

	resp, err := c.client.QueryApplicationStatus(req)
	if err != nil {
		return state, err
	}

	l.Info("request id: ", resp.RequestId)
	l.Debugf("QueryAppStatus response, appName: %s, code: %d, message: %s, app: %+v",
		appName, resp.Code, resp.Message, resp.AppInfo)

	var orderList *api.ChangeOrderList
	if orderList, err = c.ListRecentChangeOrderInfo(appID); err != nil {
		return state, errors.Wrap(err, "list recent change order info")
	}

	lastOrderType := types.CHANGE_TYPE_CREATE
	if len(orderList.ChangeOrder) > 0 {
		sort.Sort(types.ByCreateTime(orderList.ChangeOrder))
		lastOrderType = types.ChangeType(orderList.ChangeOrder[len(orderList.ChangeOrder)-1].CoType)
	}

	if len(resp.AppInfo.EccList.Ecc) == 0 {
		state = types.AppStatusStopped
	} else {
		//There may be multiple instances, as long as one is not running, return
		for _, ecc := range resp.AppInfo.EccList.Ecc {
			appState := types.AppState(ecc.AppState)
			taskState := types.TaskState(ecc.TaskState)
			if appState == types.APP_STATE_AGENT_OFF || appState == types.APP_STATE_RUNNING_BUT_URL_FAILED {
				if taskState == types.TASK_STATE_PROCESSING {
					state = types.AppStatusDeploying
				} else {
					state = types.AppStatusFailed
				}
				break
			}
			if appState == types.APP_STATE_STOPPED {
				if taskState == types.TASK_STATE_PROCESSING {
					state = types.AppStatusDeploying
				} else if taskState == types.TASK_STATE_FAILED {
					state = types.AppStatusFailed
					break
				} else if taskState == types.TASK_STATE_UNKNOWN {
					state = types.AppStatusUnknown
					break
				} else if taskState == types.TASK_STATE_SUCCESS {
					if lastOrderType != types.CHANGE_TYPE_CREATE {
						state = types.AppStatusStopped
						break
					} else {
						state = types.AppStatusDeploying
					}
				}
			}
		}
	}

	l.Infof("successfully to query app status: %v, app name: %s", state, appName)
	return state, nil
}

// InsertK8sApp Create Application
// InsertK8sApplication
// Question 1: The service name does not support "_"
func (c *wrapEDAS) InsertK8sApp(spec *types.ServiceSpec) (string, error) {
	l := c.l.WithField("func", "InsertK8sApp")

	l.Infof("start to insert app: %s", spec.Name)

	req := api.CreateInsertK8sApplicationRequest()
	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.SetDomain(c.addr)
	req.ClusterId = c.clusterID
	if len(c.logicalRegionID) != 0 {
		req.LogicalRegionId = c.logicalRegionID
	}

	req.AppName = spec.Name
	req.ImageUrl = spec.Image
	req.Command = spec.Cmd
	req.CommandArgs = spec.Args
	req.Envs = spec.Envs
	req.LocalVolume = spec.LocalVolume
	req.Liveness = spec.Liveness
	req.Readiness = spec.Readiness
	req.Annotations = spec.Annotations
	req.Labels = spec.Labels
	req.Replicas = requests.NewInteger(spec.Instances)
	if c.unLimitCPU == "true" {
		req.RequestsCpu = requests.NewInteger(spec.CPU)
		req.LimitCpu = requests.NewInteger(spec.CPU)
	} else {
		req.RequestsmCpu = requests.NewInteger(spec.Mcpu)
		req.LimitmCpu = requests.NewInteger(spec.Mcpu)
	}
	req.RequestsMem = requests.NewInteger(spec.Mem)
	req.LimitMem = requests.NewInteger(spec.Mem)

	l.Infof("insert k8s application, request body: %+v", req)

	// InsertK8sApplication
	resp, err := c.client.InsertK8sApplication(req)
	if err != nil {
		if resp != nil {
			return "", errors.Errorf("edas insert app, response http context: %s, error: %v", resp.GetHttpContentString(), err)
		}
		return "", errors.Errorf("edas insert app, error: %v", err)
	}

	l.Info("request id: ", resp.RequestId)
	l.Debugf("InsertK8sApp response, code: %d, message: %s, applicationInfo: %+v", resp.Code, resp.Message, resp.ApplicationInfo)

	l.Debugf("start loop termination status: appName: %s", req.AppName)

	// check edas app status
	if len(resp.ApplicationInfo.ChangeOrderId) != 0 {
		status, err := c.LoopTerminationStatus(resp.ApplicationInfo.ChangeOrderId)
		if err != nil {
			return "", errors.Wrapf(err, "get insert status by loop")
		}

		if status != types.CHANGE_ORDER_STATUS_SUCC {
			return "", errors.Errorf("failed to get the change order of inserting app, status: %s", types.ChangeOrderStatusString[status])
		}
	}

	l.Debugf("start loop check k8s service status: appName: %s", req.AppName)

	appID := resp.ApplicationInfo.AppId
	l.Infof("successfully to insert app name: %s, appID: %s", spec.Name, appID)
	return appID, nil
}

// DeployApp Deploy the application
// Role: The role of this interface is replace, temporarily only supports image tag update
func (c *wrapEDAS) DeployApp(appID string, spec *types.ServiceSpec) error {
	l := c.l.WithField("func", "DeployApp")

	if spec == nil {
		return errors.Errorf("invalid params: service spec is null")
	}

	l.Infof("start to deploy app, id: %s", appID)

	req := api.CreateDeployK8sApplicationRequest()
	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.SetDomain(c.addr)

	// Compatible with edas proprietary cloud version 3.7.1
	res := strings.SplitAfter(spec.Image, ":")
	splitLen := len(res)
	if splitLen > 0 {
		req.ImageTag = res[splitLen-1]
	}

	req.AppId = appID
	req.Image = spec.Image
	req.Replicas = requests.NewInteger(spec.Instances)
	req.Command = spec.Cmd
	req.Args = spec.Args
	req.Envs = spec.Envs
	req.LocalVolume = spec.LocalVolume
	req.Liveness = spec.Liveness
	req.Readiness = spec.Readiness
	req.Annotations = spec.Annotations
	req.Labels = spec.Labels
	req.Replicas = requests.NewInteger(spec.Instances)
	if c.unLimitCPU == "true" {
		req.CpuRequest = requests.NewInteger(spec.CPU)
		req.CpuLimit = requests.NewInteger(spec.CPU)
	} else {
		req.McpuRequest = requests.NewInteger(spec.Mcpu)
		req.McpuLimit = requests.NewInteger(spec.Mcpu)
	}
	req.MemoryRequest = requests.NewInteger(spec.Mem)
	req.MemoryLimit = requests.NewInteger(spec.Mem)

	// HACK: edas don't support k8s container probe
	// This value is equivalent to k8s min-ready-seconds, for coarse-grained control
	// https://kubernetes.io/docs/concepts/workloads/controllers/deployment/?spm=a2c4g.11186623.2.3.7N5Zxk#min-ready-seconds
	req.BatchWaitTime = requests.NewInteger(minReadySeconds)

	l.Infof("deploy k8s application, request body: %+v", req)

	resp, err := c.client.DeployK8sApplication(req)
	if err != nil {
		if resp != nil {
			return errors.Errorf("response http context: %s, error: %v", resp.GetHttpContentString(), err)
		}
		return errors.Errorf("error: %v", err)
	}

	l.Info("request id: ", resp.RequestId)
	l.Debugf("DeployApp response, requestID: %s, code: %d, message: %s, ChangeOrderId: %+v",
		resp.RequestId, resp.Code, resp.Message, resp.ChangeOrderId)

	if len(resp.ChangeOrderId) != 0 {
		status, err := c.LoopTerminationStatus(resp.ChangeOrderId)
		if err != nil {
			return errors.Wrapf(err, "failed to get the status of deploying app")
		}

		if status != types.CHANGE_ORDER_STATUS_SUCC {
			return errors.Errorf("failed to get the status of deploying app, status: %s", types.ChangeOrderStatusString[status])
		}
	}

	l.Debugf("successfully to deploy app, id: %s", appID)

	return nil
}

// DeleteAppByName delete application with name
func (c *wrapEDAS) DeleteAppByName(appName string) error {
	l := c.l.WithField("func", "DeleteAppByName")

	l.Infof("start to delete app: %s", appName)
	// get appId
	appID, err := c.GetAppID(appName)
	if err != nil {
		if errors.Is(err, ErrApplicationNotFound) {
			return nil
		}
		return err
	}

	orderList, err := c.ListRecentChangeOrderInfo(appID)
	if err != nil {
		l.Errorf("failed to list recent change order info, app id: %s, err: %v", appID, err)
		return err
	}

	if len(orderList.ChangeOrder) > 0 && orderList.ChangeOrder[0].Status == 1 {
		if err := c.AbortChangeOrder(orderList.ChangeOrder[0].ChangeOrderId); err != nil {
			l.Errorf("failed to abort change order, change order id: %s, err: %v", orderList.ChangeOrder[0].ChangeOrderId, err)
		}
	}

	return c.deleteAppByID(appID)
}

func (c *wrapEDAS) ScaleApp(appID string, replica int) error {
	l := c.l.WithField("func", "ScaleApp")

	l.Infof("start to scale app: %s, target replica: %d", appID, replica)

	req := api.CreateScaleK8sApplicationRequest()
	req.SetDomain(c.addr)

	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.AppId = appID
	req.RegionId = c.regionID
	req.Replicas = requests.NewInteger(replica)

	resp, err := c.client.ScaleK8sApplication(req)
	if err != nil {
		return errors.Errorf("scale k8s application err: %v", err)
	}

	l.Info("request id: ", resp.RequestId)
	l.Debugf("operation ScaleK8sApplication response, requestID: %s, code: %d, message: %s, ChangeOrderId: %+v",
		resp.RequestId, resp.Code, resp.Message, resp.ChangeOrderId)

	l.Infof("successfully to scale app id: %s", appID)
	return nil
}

func (c *wrapEDAS) GetAppDeployment(appName string) (*appsv1.Deployment, error) {
	l := c.l.WithField("func", "GetAppDeployment")

	appID, err := c.GetAppID(appName)
	if err != nil {
		return nil, err
	}

	req := api.CreateGetAppDeploymentRequest()
	req.SetDomain(c.addr)
	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.AppId = appID

	resp, err := c.client.GetAppDeployment(req)
	if err != nil {
		if resp.GetHttpStatus() == http.StatusNotFound {
			return nil, ErrAPINotSupport
		}
		return nil, errors.Wrapf(err, "get edas app %s deployment", appName)
	}

	l.Info("request id: ", resp.RequestId)

	var deployment appsv1.Deployment

	if err = json.Unmarshal([]byte(resp.Data), &deployment); err != nil {
		l.Errorf("failed to parse deployment, resp: %+v", resp.Data)
		return nil, errors.Wrap(err, "failed to parse deployment")
	}

	l.Infof("app %s get deployment, name: %s, namespace: %s", appName, deployment.Name, deployment.Namespace)

	return &deployment, nil
}
