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

package container

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/serviceutil"
	"github.com/erda-project/erda/modules/cmdb/types"
)

const (
	// EDAS edas集群
	EDAS = "edas"
)

// Container 资源对象操作封装
type Container struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 Container 对象的配置选项
type Option func(*Container)

// New 新建 Container 实例，通过 Container 实例操作企业资源
func New(options ...Option) *Container {
	o := &Container{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *Container) {
		o.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(c *Container) {
		c.bdl = bdl
	}
}

// Create 创建容器
func (c *Container) Create(container *model.Container) error {
	logrus.Debugf("container create info: %+v", container)
	return c.db.CreateContainer(container)
}

// Update 更新容器
func (c *Container) Update(container *model.Container) error {
	logrus.Debugf("container update info: %+v", container)
	return c.db.UpdateContainer(container)
}

// ListRunningByService 根据serviceName 获取运行中实例列表
func (c *Container) ListRunningByService(runtimeID, serviceName string) ([]apistructs.Container, error) {
	containers, err := c.db.ListRunningContainersByService(runtimeID, serviceName)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.convert(uniqueContainers), nil
}

// ListStoppedByService 根据serviceName 获取已停止实例列表
func (c *Container) ListStoppedByService(runtimeID, serviceName string) ([]apistructs.Container, error) {
	containers, err := c.db.ListStoppedContainersByService(runtimeID, serviceName)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.convert(uniqueContainers), nil
}

// ListAllByService 根据serviceName 获取所有实例列表(包含运行中 & 已停止)
func (c *Container) ListAllByService(runtimeID, serviceName string) ([]apistructs.Container, error) {
	containers, err := c.db.ListAllContainersByService(runtimeID, serviceName)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.convert(uniqueContainers), nil
}

// ListEdasByParams 根据 params 过滤出 EDAS 容器列表
func (c *Container) ListEdasByParams(params *apistructs.EdasContainerListRequest) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListEdasContainerByParam(params)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqueContainers), nil
}

// ListByCluster 根据 clusterName 获取容器列表
func (c *Container) ListByCluster(clusterName, orgID string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByCluster(clusterName, orgID, running)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqueContainers), nil
}

// GetContainerByTaskIDOrContainerID
func (c *Container) GetContainerByTaskIDOrContainerID(cluster, taskID, containerID string) ([]model.Container, error) {
	containers, err := c.db.GetContainerByTaskIDOrContainerID(cluster, taskID, containerID)
	if err != nil {
		return nil, err
	}
	return c.removeDuplicated(containers), nil
}

// ListByHost 根据 clusterName & hostIP 获取容器列表
func (c *Container) ListByHost(clusterName, hostIP string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByHost(clusterName, hostIP, running)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqueContainers), nil
}

// ListRunningByClusterAndHost  根据 cluster & hostIP 获取运行中容器列表
func (c *Container) ListRunningByClusterAndHost(clusterName, hostIP string) ([]apistructs.Container, error) {
	containers, err := c.db.ListRunningContainersByHost(clusterName, hostIP)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.convert(uniqueContainers), nil
}

// GetRunningAddonByClusterAndAddon 根据 clusterName 获取运行中 addon 容器列表
func (c *Container) GetRunningAddonByClusterAndAddon(clusterName, addonID string) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.GetRunningAddonByClusterAndInstanceID(clusterName, addonID)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(*containers)

	return c.parseContainers(uniqueContainers), nil
}

// GetRunningComponentByClusterAndComponent 根据 clusterName 获取运行中 component 容器列表
func (c *Container) GetRunningComponentByClusterAndComponent(clusterName, component string) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.GetRunningComponentContainersByClusterAndComponent(clusterName, component)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqueContainers), nil
}

// ListComponentUsageByCluster 根据 clusterName 获取平台组件资源使用列表
func (c *Container) ListComponentUsageByCluster(clusterName string) ([]apistructs.ComponentUsageFetchResponseData, error) {
	containers, err := c.db.ListRunningComponentContainerByCluster(clusterName)
	if err != nil {
		return nil, err
	}
	uniqueContainers := c.removeDuplicated(containers)
	// 按平台组件名称统计平台组件资源使用
	components := make(map[string]apistructs.ComponentUsageFetchResponseData, len(uniqueContainers))
	for i := range uniqueContainers {
		componentName := uniqueContainers[i].DiceComponent
		if comp, ok := components[componentName]; ok {
			comp.Instance++
			comp.CPU += uniqueContainers[i].CPU
			comp.Disk += float64(uniqueContainers[i].Disk)
			comp.Memory += float64(uniqueContainers[i].Memory)

			components[componentName] = comp
		} else {
			comp.Name = uniqueContainers[i].DiceComponent
			comp.Instance++
			comp.CPU += uniqueContainers[i].CPU
			comp.Disk += float64(uniqueContainers[i].Disk)
			comp.Memory += float64(uniqueContainers[i].Memory)

			components[componentName] = comp
		}
	}

	// 平台组件按组件名称排序
	keys := make([]string, 0, len(components))
	for k, v := range components {
		keys = append(keys, k)

		v.CPU = serviceutil.Round(v.CPU, 2)
		v.Memory = math.Ceil(v.Memory / apistructs.MB)
		v.Disk = math.Ceil(v.Disk / apistructs.MB)

		components[k] = v
	}
	sort.Strings(keys)

	usages := make([]apistructs.ComponentUsageFetchResponseData, len(keys))
	for i, k := range keys {
		usages[i] = components[k]
	}
	return usages, nil
}

// ListByOrg 根据 orgID 获取容器列表
func (c *Container) ListByOrg(orgID string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByOrg(orgID, running)
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqContainers), nil
}

// ListByProject 根据 projectID 获取容器列表
func (c *Container) ListByProject(projectID, orgID string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByProject(projectID, orgID, running)
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqContainers), nil
}

// ListByApp 根据 appID 获取容器列表
func (c *Container) ListByApp(appID, orgID string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByApplication(appID, orgID, running)
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqContainers), nil
}

// ListByRuntime 根据 runtimeID 获取容器列表
func (c *Container) ListByRuntime(runtimeID, orgID string, running bool) ([]apistructs.ContainerFetchResponseData, error) {
	containers, err := c.db.ListContainerByRuntime(runtimeID, orgID, running)
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	return c.parseContainers(uniqContainers), nil
}

// CreateOrUpdateContainer 创建或更新容器
func (c *Container) CreateOrUpdateContainer(container *model.Container) error {
	containers, err := c.GetContainerByTaskIDOrContainerID(container.Cluster, container.TaskID, container.ContainerID)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return c.Create(container)
	}

	current := &containers[0]
	latestStatus := current.Status
	for _, c := range containers {
		if types.ContainerStatusIndex[types.InstanceStatus(c.Status)] > types.ContainerStatusIndex[types.InstanceStatus(latestStatus)] {
			latestStatus = c.Status
		}
	}

	current.Status = latestStatus
	fillCmContainer(current, container)

	return c.Update(container)
}

// SyncContainerInfo 容器全量事件同步
func (c *Container) SyncContainerInfo(containers []*model.Container) error {
	containersMap := make(map[string]*model.Container, len(containers))
	for _, v := range containers {
		if v.TaskID != "" {
			containersMap[v.TaskID] = v
		}
	}

	oldContainers, err := c.db.ListRunningContainersByHost(containers[0].Cluster, containers[0].HostPrivateIPAddr)
	if err != nil {
		return err
	}
	oldContainersMap := make(map[string]*model.Container, len(oldContainers))
	for i, v := range oldContainers {
		if v.TaskID != "" {
			oldContainersMap[v.TaskID] = &oldContainers[i]
		}
	}

	for _, v := range containers {
		if oc, ok := oldContainersMap[v.TaskID]; ok {
			// 全量同步时，若container信息在事件有，在DB里也有，若DB里记录containerID为空，则更新一把
			if oc.ContainerID == "" {
				v.ID = oc.ID
				v.CreatedAt = oc.CreatedAt
				v.Status = oc.Status
				if err := c.Update(v); err != nil {
					logrus.Infof("sync container err: %v", err)
				}
			}
		} else {
			// 全量同步时，若container信息在事件里有，在DB里没有，则入库
			if err := c.Create(v); err != nil {
				logrus.Infof("sync container err: %v", err)
			}
		}
	}

	// 全量同步时，若container信息在事件里无，在DB里有，则将其置为stopped
	for _, v := range oldContainers {
		if _, ok := containersMap[v.TaskID]; !ok {
			v.Status = string(types.InstanceStatusStopped)
			if err := c.Update(&v); err != nil {
				logrus.Infof("sync container err: %v", err)
			}
		}
	}

	return nil
}

// ConvertToContainer 根据kafka推送的消息转换成contaienr对象
func (c *Container) ConvertToContainer(fields map[string]interface{}) *model.Container {
	var container model.Container
	if id, ok := fields["id"]; ok {
		container.ContainerID = id.(string)
	}
	if ip, ok := fields["ip"]; ok {
		container.IPAddress = ip.(string)
	}
	if cluster, ok := fields["cluster_name"]; ok {
		container.Cluster = cluster.(string)
	}
	if host, ok := fields["host_ip"]; ok {
		container.HostPrivateIPAddr = host.(string)
	}
	if startedAt, ok := fields["started_at"]; ok {
		container.StartedAt = startedAt.(string)
	}
	if finishedAt, ok := fields["finished_at"]; ok && !strings.Contains(finishedAt.(string), "0001-01-01T00:00:00Z") {
		container.FinishedAt = finishedAt.(string)
	}
	if image, ok := fields["image"]; ok {
		container.Image = image.(string)
	}
	if cpu, ok := fields["cpu"]; ok {
		container.CPU = dao.Round(cpu.(float64), 2)
	}
	if memory, ok := fields["memory"]; ok {
		container.Memory = (int64)(memory.(float64))
	}
	if disk, ok := fields["disk"]; ok {
		container.Disk = (int64)(disk.(float64))
	}
	if exitCode, ok := fields["exit_code"]; ok {
		container.ExitCode = (int)(exitCode.(float64))
	}
	if privileged, ok := fields["privileged"]; ok {
		container.Privileged = privileged.(bool)
	}
	if status, ok := fields["status"]; ok {
		container.Status = status.(string)
	}
	if diceOrg, ok := fields["dice_org"]; ok {
		container.DiceOrg = diceOrg.(string)
	}
	if diceProject, ok := fields["dice_project"]; ok {
		container.DiceProject = diceProject.(string)
	}
	if diceApplication, ok := fields["dice_application"]; ok {
		container.DiceApplication = diceApplication.(string)
	}
	if diceRuntime, ok := fields["dice_runtime"]; ok {
		container.DiceRuntime = diceRuntime.(string)
	}
	if diceService, ok := fields["dice_service"]; ok {
		container.DiceService = diceService.(string)
	}
	if edasAppID, ok := fields["edas_app_id"]; ok {
		container.EdasAppID = edasAppID.(string)
	}
	if edasAppName, ok := fields["edas_app_name"]; ok {
		container.EdasAppName = edasAppName.(string)
	}
	if edasGroupID, ok := fields["edas_group_id"]; ok {
		container.EdasGroupID = edasGroupID.(string)
	}
	if diceProjectName, ok := fields["dice_project_name"]; ok {
		container.DiceProjectName = diceProjectName.(string)
	}
	if diceApplicationName, ok := fields["dice_application_name"]; ok {
		container.DiceApplicationName = diceApplicationName.(string)
	}
	if diceRuntimeName, ok := fields["dice_runtime_name"]; ok {
		container.DiceRuntimeName = diceRuntimeName.(string)
	}
	if diceComponent, ok := fields["dice_component"]; ok {
		container.DiceComponent = diceComponent.(string)
	}
	if diceAddon, ok := fields["dice_addon"]; ok {
		container.DiceAddon = diceAddon.(string)
	}
	if diceAddonName, ok := fields["dice_addon_name"]; ok {
		container.DiceAddonName = diceAddonName.(string)
	}
	if diceWorkspace, ok := fields["dice_workspace"]; ok {
		container.DiceWorkspace = diceWorkspace.(string)
	}
	if diceSharedLevel, ok := fields["dice_shared_level"]; ok {
		container.DiceSharedLevel = diceSharedLevel.(string)
	}
	if taskID, ok := fields["task_id"]; ok {
		container.TaskID = taskID.(string)
	}

	return &container
}

func fillCmContainer(oldC, newC *model.Container) {
	newC.ID = oldC.ID
	logrus.Debugf("origin container: %+v", *newC)

	if len(oldC.ContainerID) != 0 && len(newC.ContainerID) == 0 {
		newC.ContainerID = oldC.ContainerID
	}
	if len(oldC.StartedAt) != 0 && len(newC.StartedAt) == 0 {
		newC.StartedAt = oldC.StartedAt
	}
	if len(oldC.FinishedAt) != 0 && !strings.Contains(oldC.FinishedAt, "0001-01-01T00:00:00Z") {
		newC.FinishedAt = oldC.FinishedAt
	}
	if newC.ExitCode != oldC.ExitCode && newC.ExitCode == 0 {
		newC.ExitCode = oldC.ExitCode
	}
	if oldC.Privileged {
		newC.Privileged = oldC.Privileged
	}
	if len(oldC.Cluster) != 0 && len(newC.Cluster) == 0 {
		newC.Cluster = oldC.Cluster
	}
	if len(oldC.HostPrivateIPAddr) != 0 && len(newC.HostPrivateIPAddr) == 0 {
		newC.HostPrivateIPAddr = oldC.HostPrivateIPAddr
	}
	if len(oldC.IPAddress) != 0 && len(newC.IPAddress) == 0 {
		newC.IPAddress = oldC.IPAddress
	}
	if len(oldC.Image) != 0 && len(newC.Image) == 0 {
		newC.Image = oldC.Image
	}
	if oldC.CPU != 0 {
		newC.CPU = oldC.CPU
	}
	if oldC.Memory != 0 {
		newC.Memory = oldC.Memory
	}
	if oldC.Disk != 0 {
		newC.Disk = oldC.Disk
	}
	if len(oldC.DiceOrg) != 0 {
		newC.DiceOrg = oldC.DiceOrg
	}
	if len(oldC.DiceProject) != 0 {
		newC.DiceProject = oldC.DiceProject
	}
	if len(oldC.DiceApplication) != 0 {
		newC.DiceApplication = oldC.DiceApplication
	}
	if len(oldC.DiceRuntime) != 0 {
		newC.DiceRuntime = oldC.DiceRuntime
	}
	if len(oldC.DiceService) != 0 {
		newC.DiceService = oldC.DiceService
	}
	if len(oldC.DiceProjectName) != 0 {
		newC.DiceProjectName = oldC.DiceProjectName
	}
	if len(oldC.DiceApplicationName) != 0 {
		newC.DiceApplicationName = oldC.DiceApplicationName
	}
	if len(oldC.DiceRuntimeName) != 0 {
		newC.DiceRuntimeName = oldC.DiceRuntimeName
	}
	if len(oldC.DiceComponent) != 0 {
		newC.DiceComponent = oldC.DiceComponent
	}
	if len(oldC.DiceAddon) != 0 {
		newC.DiceAddon = oldC.DiceAddon
	}
	if len(oldC.DiceAddonName) != 0 {
		newC.DiceAddonName = oldC.DiceAddonName
	}
	if len(oldC.DiceWorkspace) != 0 {
		newC.DiceWorkspace = oldC.DiceWorkspace
	}
	if len(oldC.DiceSharedLevel) != 0 {
		newC.DiceSharedLevel = oldC.DiceSharedLevel
	}
	if isUpdateInstanceStatus(oldC.Status, newC.Status) {
		newC.Status = oldC.Status
	}
	if oldC.TimeStamp != 0 {
		newC.TimeStamp = oldC.TimeStamp
	}
	if len(oldC.TaskID) != 0 && len(newC.TaskID) == 0 {
		newC.TaskID = oldC.TaskID
	}
	newC.CreatedAt = oldC.CreatedAt
	newC.UpdatedAt = time.Now()
}

// 判断是否使用之前的容器状态进行更新，不可逆(除去 Healthy & Unhealthy)
func isUpdateInstanceStatus(oldStatus, newStatus string) bool {
	if types.ContainerStatusIndex[types.InstanceStatus(oldStatus)] == types.ContainerStatusIndex[types.InstanceStatus(newStatus)] &&
		(newStatus == string(types.InstanceStatusHealthy) || newStatus == string(types.InstanceStatusUnHealthy)) {
		return false
	}

	if types.ContainerStatusIndex[types.InstanceStatus(oldStatus)] > types.ContainerStatusIndex[types.InstanceStatus(newStatus)] {
		return true
	}

	return false
}

// SearchService 服务范围内搜索
// TODO 专为搜索使用，搜索API重构后可去除
func (c *Container) SearchService(clusterName, keyword, orgID string) ([]apistructs.Resource, error) {
	var (
		result     []apistructs.Resource
		containers []model.Container
		err        error
	)

	containers, err = c.db.ListRunningProjectContainersByCluster(clusterName, orgID)
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	projects := make(map[string]*apistructs.ProjectCache)
	applications := make(map[string]*apistructs.ApplicationUsageFetchResponseData)
	runtimes := make(map[string]*apistructs.RuntimeUsageFetchResponseData)
	servs := make(map[string]*apistructs.ServiceUsageFetchResponseData)

	for _, v := range uniqContainers {
		if c.containsKeyword(v, keyword) {
			c.sortOutServices(&c.parseContainers([]model.Container{v})[0], projects, applications, runtimes, servs)
		}
	}

	result = c.serviceMerge(projects, applications, runtimes, servs)

	return result, nil
}

// SearchContainer 按容器过滤
// TODO 搜索API重构后可去除
func (c *Container) SearchContainer(cluster, keyword, orgID string) ([]apistructs.Resource, error) {
	result, err := c.SearchService(cluster, keyword, orgID)
	if err != nil {
		return nil, err
	}

	resultComponent, err := c.ComponentOrAddonSearch(cluster, "component", keyword)
	if err != nil {
		return nil, err
	}

	resultAddon, err := c.ComponentOrAddonSearch(cluster, "addon", keyword)
	if err != nil {
		return nil, err
	}

	result = append(result, resultComponent...)
	result = append(result, resultAddon...)

	return result, nil
}

// ComponentOrAddonSearch 按compoenent & addon 过滤
// TODO 专为搜索使用，搜索API重构后可去除
func (c *Container) ComponentOrAddonSearch(cluster, searchType, searchValue string) ([]apistructs.Resource, error) {
	var (
		containers []model.Container
		err        error
	)

	switch searchType {
	case apistructs.ComponentSearchType:
		containers, err = c.db.ListRunningComponentContainerByCluster(cluster)
	case apistructs.AddonSearchType:
		containers, err = c.db.ListRunningAddonsByCluster(cluster)
	}
	if err != nil {
		return nil, err
	}
	uniqContainers := c.removeDuplicated(containers)

	extras := make(map[string]apistructs.ExtraResource, 10)
	result := make([]apistructs.Resource, 0)

	for _, container := range uniqContainers {
		var cName string
		if searchType == apistructs.AddonSearchType {
			cName = container.DiceAddonName
		} else {
			cName = container.DiceComponent
		}
		if cName != "" && c.containsKeyword(container, searchValue) {
			c.extraMerge(c.parseContainers([]model.Container{container})[0], extras, searchType, cName)
		}
	}

	for _, extra := range extras {
		resource := apistructs.Resource{
			Type:     extra.Type,
			Resource: extra,
		}
		result = append(result, resource)
	}
	return result, nil
}

// 去除taskID/containerID重复的容器实例记录
func (c *Container) removeDuplicated(containers []model.Container) []model.Container {
	containerMap := make(map[string]*model.Container, len(containers))

	for i := range containers {
		if containers[i].TaskID == "" && containers[i].ContainerID == "" { // 实例记录理论不应该出现taskID/containerID同时为空的情况
			continue
		}
		var key string
		if containers[i].TaskID == "" {
			key = containers[i].ContainerID
		} else {
			key = containers[i].TaskID
		}

		v, ok := containerMap[key]
		if !ok {
			containerMap[key] = &containers[i]
		} else {
			// 若存在key相同的多条记录, 且状态推进更靠后，则使用新的状态
			if types.ContainerStatusIndex[types.InstanceStatus(containers[i].Status)] >
				types.ContainerStatusIndex[types.InstanceStatus(v.Status)] {
				containerMap[key].Status = containers[i].Status
			}
			// 若containerID不为空，则更新至已有记录
			if v.ContainerID == "" && containers[i].ContainerID != "" {
				v.ContainerID = containers[i].ContainerID
			}
		}
	}
	newContainers := make(ContainerList, 0, len(containers))
	for _, v := range containerMap {
		v.CPU = serviceutil.Round(v.CPU, 2)
		newContainers = append(newContainers, *v)
	}
	sort.Sort(sort.Reverse(newContainers))

	return newContainers
}

func (c *Container) convert(containers []model.Container) []apistructs.Container {
	result := make([]apistructs.Container, 0, len(containers))
	for i := range containers {
		item := apistructs.Container{
			ID:          containers[i].TaskID,
			ContainerID: containers[i].ContainerID,
			IPAddress:   containers[i].IPAddress,
			Host:        containers[i].HostPrivateIPAddr,
			Image:       containers[i].Image,
			CPU:         containers[i].CPU,
			Memory:      containers[i].Memory,
			Disk:        containers[i].Disk,
			StartedAt:   containers[i].StartedAt,
			UpdatedAt:   containers[i].FinishedAt,
			Status:      c.parseStatus(containers[i].Cluster, containers[i].Status),
			Service:     containers[i].DiceService,
		}
		result = append(result, item)
	}

	return result
}

func (c *Container) parseStatus(clusterName, status string) string {
	// instanceStatusRunning 是为了兼容老数据
	if status == string(types.InstanceStatusRunning) {
		status = string(types.InstanceStatusStarting)
	}
	// TODO: edas 集群的特殊处理，暂无实例健康信息
	if status == string(types.InstanceStatusStarting) {
		cluster, err := c.db.GetClusterByName(clusterName)
		if err != nil {
			logrus.Warnf("failed to fetch cluster: %v", err)
		}
		if cluster != nil && cluster.Type == apistructs.EDAS {
			status = string(types.InstanceStatusHealthy)
		}
	}

	return status
}

// TODO 后续考虑 apistructs.Container & apistructs.ContainerFetchResponseData 合并，暂兼容保留
func (c *Container) parseContainers(containers []model.Container) []apistructs.ContainerFetchResponseData {
	result := make([]apistructs.ContainerFetchResponseData, 0, len(containers))
	for _, item := range containers {
		container := apistructs.ContainerFetchResponseData{
			ID:                  item.ContainerID,
			Deleted:             item.Deleted,
			StartedAt:           item.StartedAt,
			FinishedAt:          item.FinishedAt,
			ExitCode:            item.ExitCode,
			Privileged:          item.Privileged,
			Cluster:             item.Cluster,
			HostPrivateIPAddr:   item.HostPrivateIPAddr,
			IPAddress:           item.IPAddress,
			Image:               item.Image,
			CPU:                 item.CPU,
			Memory:              item.Memory,
			Disk:                item.Disk,
			DiceOrg:             item.DiceOrg,
			DiceProject:         item.DiceProject,
			DiceApplication:     item.DiceApplication,
			DiceRuntime:         item.DiceRuntime,
			DiceService:         item.DiceService,
			EdasAppID:           item.EdasAppID,
			EdasAppName:         item.EdasAppName,
			EdasGroupID:         item.EdasGroupID,
			DiceProjectName:     item.DiceProjectName,
			DiceApplicationName: item.DiceApplicationName,
			DiceRuntimeName:     item.DiceRuntimeName,
			DiceComponent:       item.DiceComponent,
			DiceAddon:           item.DiceAddon,
			DiceAddonName:       item.DiceAddonName,
			DiceWorkspace:       item.DiceWorkspace,
			DiceSharedLevel:     item.DiceSharedLevel,
			Status:              item.Status,
			TimeStamp:           item.TimeStamp,
			TaskID:              item.TaskID,
			Env:                 item.Env,
		}
		result = append(result, container)
	}

	return result
}

// TODO 专为搜索使用，搜索API重构后可去除
func (c *Container) extraMerge(container apistructs.ContainerFetchResponseData, extras map[string]apistructs.ExtraResource, extraType, name string) {
	var extra apistructs.ExtraResource
	if ex, ok := extras[name]; ok {
		ex.Resource = append(ex.Resource, container)
		ex.Usage.Memory += float64(container.Memory)
		ex.Usage.Disk += float64(container.Disk)
		ex.Usage.CPU += container.CPU
		ex.Usage.Instance++
		extras[name] = ex
	} else {
		extra.Type = extraType
		extra.Resource = append(extra.Resource, container)
		extra.Usage.Instance = 1
		extra.Usage.Name = name
		extra.Usage.Memory = float64(container.Memory)
		extra.Usage.Disk = float64(container.Disk)
		extra.Usage.CPU = container.CPU
		extras[name] = extra
	}
}

// TODO 专为搜索使用，搜索API重构后可去除
func (c *Container) serviceMerge(projects map[string]*apistructs.ProjectCache,
	applications map[string]*apistructs.ApplicationUsageFetchResponseData,
	runtimes map[string]*apistructs.RuntimeUsageFetchResponseData,
	services map[string]*apistructs.ServiceUsageFetchResponseData) []apistructs.Resource {
	result := make([]apistructs.Resource, 0)

	for name, pro := range projects {
		var serviceResource apistructs.ServiceResource
		var resource apistructs.Resource

		serviceResource.Name = name
		serviceResource.Resource = pro.Resource
		serviceResource.ProjectUsage = pro.Usage

		for fullName := range pro.Application {
			serviceResource.ApplicationUsage = append(serviceResource.ApplicationUsage, applications[fullName])
		}

		for fullName := range pro.Runtime {
			serviceResource.RuntimeUsage = append(serviceResource.RuntimeUsage, runtimes[fullName])
		}

		for fullName := range pro.Services {
			serviceResource.ServiceUsage = append(serviceResource.ServiceUsage, services[fullName])
		}
		resource.Type = apistructs.ServiceSearchType
		resource.Resource = serviceResource
		result = append(result, resource)
	}
	return result
}

// TODO 专为搜索使用，搜索API重构后可去除
func (c *Container) containsKeyword(container model.Container, keyword string) bool {
	splitStr := "/"
	containerStr := container.DiceProject + splitStr +
		container.DiceProjectName + splitStr +
		container.DiceApplication + splitStr +
		container.DiceApplicationName + splitStr +
		container.DiceRuntime + splitStr +
		container.DiceRuntimeName + splitStr +
		container.DiceService + splitStr +
		container.DiceComponent + splitStr +
		container.DiceAddonName + splitStr +
		container.ContainerID + splitStr +
		container.HostPrivateIPAddr + splitStr +
		container.Image + splitStr +
		container.IPAddress

	return strings.Contains(containerStr, keyword)
}

// TODO 专为搜索添加，后续重构后可去除
func (c *Container) sortOutServices(container *apistructs.ContainerFetchResponseData,
	projects map[string]*apistructs.ProjectCache,
	applications map[string]*apistructs.ApplicationUsageFetchResponseData,
	runtimes map[string]*apistructs.RuntimeUsageFetchResponseData,
	services map[string]*apistructs.ServiceUsageFetchResponseData) {
	var appFullName, runtimeFullName, serviceFullName string

	name := container.DiceProjectName

	if ser, ok := projects[name]; ok {
		ser.Resource = append(ser.Resource, container)
		ser.Usage.Instance++
		ser.Usage.CPU += container.CPU
		ser.Usage.Disk += float64(container.Disk)
		ser.Usage.Memory += float64(container.Memory)

		appFullName = name + "/" + container.DiceApplicationName

		if _, ok := ser.Application[appFullName]; ok {
			app := applications[appFullName]
			app.Instance++
			app.CPU += container.CPU
			app.Disk += float64(container.Disk)
			app.Memory += float64(container.Memory)
		} else {
			var app apistructs.ApplicationUsageFetchResponseData
			ser.Application[appFullName] = struct{}{}
			applications[appFullName] = &app
			app.ID = container.DiceApplication
			app.Name = container.DiceApplicationName
			app.Instance++
			app.CPU += container.CPU
			app.Disk += float64(container.Disk)
			app.Memory += float64(container.Memory)
		}

		runtimeFullName = appFullName + "/" + container.DiceRuntimeName

		if _, ok := ser.Runtime[runtimeFullName]; ok {
			runtime := runtimes[runtimeFullName]
			runtime.Instance++
			runtime.CPU += container.CPU
			runtime.Disk += float64(container.Disk)
			runtime.Memory += float64(container.Memory)
		} else {
			var runtime apistructs.RuntimeUsageFetchResponseData
			ser.Runtime[runtimeFullName] = struct{}{}
			runtimes[runtimeFullName] = &runtime
			runtime.ID = container.DiceRuntime
			runtime.Name = container.DiceRuntimeName
			runtime.Application = container.DiceApplication
			runtime.Instance++
			runtime.CPU += container.CPU
			runtime.Disk += float64(container.Disk)
			runtime.Memory += float64(container.Memory)
		}

		serviceFullName = runtimeFullName + "/" + container.DiceService

		if _, ok := ser.Services[serviceFullName]; ok {
			service := services[serviceFullName]
			service.Instance++
			service.CPU += container.CPU
			service.Disk += float64(container.Disk)
			service.Memory += float64(container.Memory)
		} else {
			var service apistructs.ServiceUsageFetchResponseData
			ser.Services[serviceFullName] = struct{}{}
			services[serviceFullName] = &service
			service.Name = container.DiceService
			service.Runtime = container.DiceRuntime
			service.Instance++
			service.CPU += container.CPU
			service.Disk += float64(container.Disk)
			service.Memory += float64(container.Memory)
		}
	} else {
		var proCache apistructs.ProjectCache
		projects[name] = &proCache
		proCache.Resource = make([]*apistructs.ContainerFetchResponseData, 0)
		proCache.Resource = append(proCache.Resource, container)
		proCache.Usage = &apistructs.ProjectUsageFetchResponseData{
			Instance: 1,
			CPU:      container.CPU,
			Disk:     float64(container.Disk),
			Memory:   float64(container.Memory),
			ID:       container.DiceProject,
			Name:     container.DiceProjectName,
		}

		appFullName = name + "/" + container.DiceApplicationName

		var app apistructs.ApplicationUsageFetchResponseData
		proCache.Application = make(map[string]interface{})
		proCache.Application[appFullName] = struct{}{}
		applications[appFullName] = &app
		app.ID = container.DiceApplication
		app.Name = container.DiceApplicationName
		app.Instance++
		app.CPU += container.CPU
		app.Disk += float64(container.Disk)
		app.Memory += float64(container.Memory)

		runtimeFullName = appFullName + "/" + container.DiceRuntimeName

		var runtime apistructs.RuntimeUsageFetchResponseData
		proCache.Runtime = make(map[string]interface{})
		proCache.Runtime[runtimeFullName] = struct{}{}
		runtimes[runtimeFullName] = &runtime
		runtime.ID = container.DiceRuntime
		runtime.Name = container.DiceRuntimeName
		runtime.Application = container.DiceApplication
		runtime.Instance++
		runtime.CPU += container.CPU
		runtime.Disk += float64(container.Disk)
		runtime.Memory += float64(container.Memory)

		serviceFullName = runtimeFullName + "/" + container.DiceService

		var service apistructs.ServiceUsageFetchResponseData
		proCache.Services = make(map[string]interface{})
		proCache.Services[serviceFullName] = struct{}{}
		services[serviceFullName] = &service
		service.Name = container.DiceService
		service.Runtime = container.DiceRuntime
		service.Instance++
		service.CPU += container.CPU
		service.Disk += float64(container.Disk)
		service.Memory += float64(container.Memory)
	}
}

// ListClusterServicesList 获取指定集群的所有service列表
func (c *Container) ListClusterServices(clusterName string) ([]apistructs.ServiceUsageData, error) {
	var servicesList = []apistructs.ServiceUsageData{}

	// list application service
	containers, err := c.db.ListClusterRunningContainers(clusterName)
	if err != nil {
		return nil, err
	}
	servicesList = append(servicesList, c.composeServiceResource(containers, "application")...)

	// list addon service
	containers, err = c.db.ListRunningAddonsByCluster(clusterName)
	if err != nil {
		return nil, err
	}
	servicesList = append(servicesList, c.composeServiceResource(containers, "addon")...)

	return servicesList, nil
}

func (c *Container) composeServiceResource(containers []model.Container, serviceType string) []apistructs.ServiceUsageData {
	uniqContainers := c.removeDuplicated(containers)
	services := make(map[string]apistructs.ServiceUsageData, len(uniqContainers))
	for i := range uniqContainers {
		var serviceName string
		switch serviceType {
		case "addon":
			serviceName = uniqContainers[i].DiceAddon
		case "application":
			serviceName = uniqContainers[i].DiceService
		}

		if s, ok := services[serviceName]; ok {
			s.Instance++
			s.CPU += uniqContainers[i].CPU
			s.Disk += float64(uniqContainers[i].Disk)
			s.Memory += float64(uniqContainers[i].Memory)

			services[serviceName] = s
		} else {
			s.Instance++
			s.CPU += uniqContainers[i].CPU
			s.Disk += float64(uniqContainers[i].Disk)
			s.Memory += float64(uniqContainers[i].Memory)
			s.Project = uniqContainers[i].DiceProjectName
			s.Application = uniqContainers[i].DiceApplicationName
			s.Workspace = uniqContainers[i].DiceWorkspace
			s.Type = serviceType

			switch serviceType {
			case "addon":
				s.Name = uniqContainers[i].DiceAddonName
				s.ID = uniqContainers[i].DiceAddon
				s.SharedLevel = uniqContainers[i].DiceSharedLevel
			case "application":
				s.Name = uniqContainers[i].DiceService
			}

			services[serviceName] = s
		}
	}

	usages := make([]apistructs.ServiceUsageData, 0, len(services))
	for _, v := range services {
		v.CPU = serviceutil.Round(v.CPU, 2)
		v.Memory = math.Ceil(v.Memory / apistructs.MB)
		v.Disk = math.Ceil(v.Disk / apistructs.MB)
		usages = append(usages, v)
	}

	return usages
}

// A slice of Container that implements sort.Interface to sort by Value.
type ContainerList []model.Container

func (cl ContainerList) Len() int           { return len(cl) }
func (cl ContainerList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl ContainerList) Less(i, j int) bool { return cl[i].CreatedAt.Unix() < cl[j].CreatedAt.Unix() }
