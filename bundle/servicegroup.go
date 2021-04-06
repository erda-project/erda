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

package bundle

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateServiceGroup create servicegroup
func (b *Bundle) CreateServiceGroup(sg apistructs.ServiceGroupCreateV2Request) error {
	var resp apistructs.ServiceGroupCreateV2Response
	if err := callScheduler(b, sg, &resp, "/api/servicegroup", b.hc.Post); err != nil {
		return err
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// DeleteServiceGroup delete servicegroup
func (b *Bundle) DeleteServiceGroup(namespace, name string) error {
	host, err := b.urls.Scheduler()
	if err != nil {
		return err
	}
	var resp apistructs.ServiceGroupDeleteV2Response
	r, err := b.hc.Delete(host).Path("/api/servicegroup").
		Param("name", name).Param("namespace", namespace).
		Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("statuscode: %d", r.StatusCode()))
	}
	if !resp.Success && resp.Error.Code == "404" {
		logrus.Errorf("deleteservicegroup: %s/%s not found", namespace, name)
		return nil
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// ForceDeleteServiceGroup delete service group with force option
func (b *Bundle) ForceDeleteServiceGroup(req apistructs.ServiceGroupDeleteRequest) error {
	var force string
	if req.Force {
		force = "true"
	} else {
		force = "false"
	}
	host, err := b.urls.Scheduler()
	if err != nil {
		return err
	}
	var resp apistructs.ServiceGroupDeleteV2Response
	r, err := b.hc.Delete(host).Path("/api/servicegroup").
		Param("name", req.Name).Param("namespace", req.Namespace).Param("force", force).
		Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("statuscode: %d", r.StatusCode()))
	}
	if !resp.Success && resp.Error.Code == "404" {
		logrus.Errorf("deleteservicegroup: %s/%s not found", req.Namespace, req.Name)
		return nil
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// UpdateServiceGroup update servicegroup
func (b *Bundle) UpdateServiceGroup(sg apistructs.ServiceGroupUpdateV2Request) error {
	var resp apistructs.ServiceGroupUpdateV2Response
	if err := callScheduler(b, sg, &resp, "/api/servicegroup", b.hc.Put); err != nil {
		return err
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// InspectServiceGroup get servicegroup info
func (b *Bundle) InspectServiceGroup(namespace, name string) (
	*apistructs.ServiceGroup, error) {
	sg := apistructs.ServiceGroupInfoRequest{Type: namespace, ID: name}
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	var resp apistructs.ServiceGroupInfoResponse
	r, err := b.hc.Get(host).Path("/api/servicegroup").Param("namespace", sg.Type).Param("name", sg.ID).Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return nil, toAPIError(r.StatusCode(), resp.Error)
	}
	return &resp.Data, nil
}

// CreateJobVolume create job volume
func (b *Bundle) CreateJobVolume(v apistructs.JobVolume) (string, error) {
	var resp apistructs.JobVolumeCreateResponse
	if err := callScheduler(b, v, &resp, "/api/jobvolume", b.hc.Post); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	}
	return resp.ID, nil
}

func (b *Bundle) KillPod(r apistructs.ServiceGroupKillPodRequest) error {
	var resp apistructs.ServiceGroupKillPodResponse
	if err := callScheduler(b, r, &resp, "/api/servicegroup/actions/killpod", b.hc.Post); err != nil {
		return err
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// InspectServiceGroup get servicegroup info
func (b *Bundle) ServiceGroupConfigUpdate(sg apistructs.ServiceGroup) error {
	var resp apistructs.ServiceGroupConfigUpdateResponse
	if err := callScheduler(b, sg, &resp, "/api/servicegroup/actions/config", b.hc.Put); err != nil {
		return err
	}
	if !resp.Success {
		return toAPIError(200, resp.Error)
	}
	return nil
}

// TODO: an ugly hack, need refactor, it may cause goroutine explosion
func (b *Bundle) InspectServiceGroupWithTimeout(namespace, name string) (*apistructs.ServiceGroup, error) {
	var (
		sg  *apistructs.ServiceGroup
		err error
	)
	done := make(chan struct{}, 1)
	go func() {
		sg, err = b.InspectServiceGroup(namespace, name)
		done <- struct{}{}
	}()
	select {
	case <-done:
		return sg, err
	case <-time.After(3 * time.Second):
		return nil, apierrors.ErrInvoke.InternalError(fmt.Errorf("timeout for invoke getServiceGroup"))
	}
}

// CancelServiceGroup will be DEPRECATED
func (b *Bundle) CancelServiceGroup(namespace, name string) error {
	host, err := b.urls.Scheduler()
	if err != nil {
		return err
	}
	var body bytes.Buffer
	r, err := b.hc.Post(host).Path(fmt.Sprintf("/v1/runtime/%s/%s/cancel", namespace, name)).Do().Body(&body)
	if err != nil {
		return err
	}
	if !r.IsOK() {
		return fmt.Errorf("failed to cancel servicegroup: namespace:%s, name:%s, statusCode: %d, body: %v",
			namespace, name, r.StatusCode(), body.String())
	}
	return nil
}

// GetServiceGroupStatus will be DEPRECATED
func (b *Bundle) GetServiceGroupStatus(namespace, name string) (*apistructs.MultiLevelStatus, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	var resp apistructs.MultiLevelStatus
	r, err := b.hc.Get(host).Path(fmt.Sprintf("/v1/runtimeinfo/status/%s/%s", namespace, name)).Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get status detail of servicegroup, namespace:%s, name:%s, status-code %d",
			namespace, name, r.StatusCode())
	}
	return &resp, nil
}

func (b *Bundle) GetPodInfo(req apistructs.PodInfoRequest) (*apistructs.PodInfoResponse, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	params := make(url.Values)
	if req.Cluster != "" {
		params.Add("cluster", req.Cluster)
	}
	if req.OrgName != "" {
		params.Add("orgName", req.OrgName)
	}
	if req.OrgID != "" {
		params.Add("orgID", req.OrgID)
	}
	if req.ProjectName != "" {
		params.Add("projectName", req.ProjectName)
	}
	if req.ProjectID != "" {
		params.Add("projectID", req.ProjectID)
	}
	if req.ApplicationName != "" {
		params.Add("applicationName", req.ApplicationName)
	}
	if req.ApplicationID != "" {
		params.Add("applicationID", req.ApplicationID)
	}
	if req.RuntimeName != "" {
		params.Add("runtimeName", req.RuntimeName)
	}
	if req.RuntimeID != "" {
		params.Add("runtimeID", req.RuntimeID)
	}
	if req.ServiceName != "" {
		params.Add("serviceName", req.ServiceName)
	}
	if req.Workspace != "" {
		params.Add("workspace", req.Workspace)
	}
	if req.ServiceType != "" {
		params.Add("serviceType", req.ServiceType)
	}
	if req.AddonID != "" {
		params.Add("addonID", req.AddonID)
	}
	if len(req.Phases) != 0 {
		params.Add("phases", strutil.Join(req.Phases, ",", true))
	}
	if req.Limit != 0 {
		params.Add("limit", strconv.Itoa(req.Limit))
	}
	var resp apistructs.PodInfoResponse
	r, err := b.hc.Get(host).Path(fmt.Sprintf("/api/podinfo")).Params(params).Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get instanceinfo: req: %+v, statuscode: %d", req, r.StatusCode())
	}
	return &resp, nil
}

// GetInstanceInfo 实例状态 list
func (b *Bundle) GetInstanceInfo(req apistructs.InstanceInfoRequest) (*apistructs.InstanceInfoResponse, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	params := make(url.Values)
	if req.Cluster != "" {
		params.Add("cluster", req.Cluster)
	}
	if req.OrgName != "" {
		params.Add("orgName", req.OrgName)
	}
	if req.OrgID != "" {
		params.Add("orgID", req.OrgID)
	}
	if req.ProjectName != "" {
		params.Add("projectName", req.ProjectName)
	}
	if req.ProjectID != "" {
		params.Add("projectID", req.ProjectID)
	}
	if req.ApplicationName != "" {
		params.Add("applicationName", req.ApplicationName)
	}
	if req.ApplicationID != "" {
		params.Add("applicationID", req.ApplicationID)
	}
	if req.RuntimeName != "" {
		params.Add("runtimeName", req.RuntimeName)
	}
	if req.RuntimeID != "" {
		params.Add("runtimeID", req.RuntimeID)
	}
	if req.ServiceName != "" {
		params.Add("serviceName", req.ServiceName)
	}
	if req.Workspace != "" {
		params.Add("workspace", req.Workspace)
	}
	if req.ContainerID != "" {
		params.Add("containerID", req.ContainerID)
	}
	if req.ServiceType != "" {
		params.Add("serviceType", req.ServiceType)
	}
	if req.AddonID != "" {
		params.Add("addonID", req.AddonID)
	}
	if req.InstanceIP != "" {
		params.Add("instanceIP", req.InstanceIP)
	}
	if req.HostIP != "" {
		params.Add("hostIP", req.HostIP)
	}
	if len(req.Phases) != 0 {
		params.Add("phases", strutil.Join(req.Phases, ",", true))
	}
	if req.Limit != 0 {
		params.Add("limit", strconv.Itoa(req.Limit))
	}
	var resp apistructs.InstanceInfoResponse
	r, err := b.hc.Get(host).Path(fmt.Sprintf("/api/instanceinfo")).Params(params).Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get instanceinfo: req: %+v, statuscode: %d", req, r.StatusCode())
	}
	return &resp, nil
}

func (b *Bundle) CapacityInfo(clustername string) (*apistructs.CapacityInfoResponse, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	params := make(url.Values)
	params.Add("clusterName", clustername)
	var resp apistructs.CapacityInfoResponse
	r, err := b.hc.Get(host).Path("/api/capacity").Params(params).Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get capacity info, statuscode: %d", r.StatusCode())
	}
	return &resp, nil
}

// PARAM: brief, 是否需要已使用资源信息
func (b *Bundle) ResourceInfo(clustername string, brief bool) (*apistructs.ClusterResourceInfoData, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	params := make(url.Values)
	if brief {
		params.Add("brief", "true")
	}
	var resp apistructs.ClusterResourceInfoResponse
	r, err := b.hc.Get(host).
		Path(fmt.Sprintf("/api/resourceinfo/%s", clustername)).
		Params(params).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get resource info, statuscode: %d", r.StatusCode())
	}
	return &resp.Data, nil
}
func (b *Bundle) PrecheckServiceGroup(sg apistructs.ServiceGroupPrecheckRequest) (
	*apistructs.ServiceGroupPrecheckData, error) {
	var resp apistructs.ServiceGroupPrecheckResponse
	if err := callScheduler(b, sg, &resp, "/api/servicegroup/actions/precheck", b.hc.Post); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, toAPIError(200, resp.Error)
	}
	return &resp.Data, nil
}

func callScheduler(b *Bundle, req, resp interface{}, path string,
	httpfunc func(host string, retry ...httpclient.RetryOption) *httpclient.Request) error {
	host, err := b.urls.Scheduler()
	if err != nil {
		return err
	}
	r, err := httpfunc(host).Path(path).JSONBody(req).Do().JSON(resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("statuscode: %d, %v", r.StatusCode(), string(r.Body())))
	}
	return nil
}
