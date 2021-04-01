package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/types"
)

var allRunningStatus = []types.InstanceStatus{
	types.InstanceStatusStarting,
	types.InstanceStatusRunning,
	types.InstanceStatusHealthy,
	types.InstanceStatusUnHealthy,
	types.InstanceStatusUnknown,
}

var allStoppedStatus = []types.InstanceStatus{
	types.InstanceStatusFailed,
	types.InstanceStatusFinished,
	types.InstanceStatusKilled,
	types.InstanceStatusStopped,
	types.InstanceStatusOOM,
	types.InstanceStatusUnknown,
}

// CreateContainer 创建容器
func (client *DBClient) CreateContainer(container *model.Container) error {
	return client.Create(container).Error
}

// UpdateContainer 更新容器
func (client *DBClient) UpdateContainer(container *model.Container) error {
	return client.Save(container).Error
}

// ListContainerByCluster 根据 clusterName 获取容器列表
func (client *DBClient) ListContainerByCluster(clusterName, orgID string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("cluster = ?", clusterName).Where("dice_org = ?", orgID)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListContainerByHost 根据 clusterName & hostIP 获取容器列表
func (client *DBClient) ListContainerByHost(clusterName, hostIP string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("cluster = ?", clusterName).Where("host_private_ip_addr = ?", hostIP)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListRunningProjectContainersByCluster 根据 clusterName 获取集群内正在运行的 containers
func (client *DBClient) ListRunningProjectContainersByCluster(clusterName, orgID string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).
		Where("status in (?)", allRunningStatus).
		Where("dice_org = ?", orgID).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

// ListRunningComponentContainerByCluster 根据 clusterName 获取集群内正在运行的平台组件容器
func (client *DBClient) ListRunningComponentContainerByCluster(clusterName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).Where("status in (?)", allRunningStatus).
		Where("dice_component is not null AND dice_component <> ''").
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// GetRunningComponentContainersByClusterAndComponent 根据 clusterName & component 获取集群内正在运行的平台组件容器
func (client *DBClient) GetRunningComponentContainersByClusterAndComponent(clusterName, componentName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).
		Where("dice_component = ?", componentName).
		Where("status in (?)", allRunningStatus).
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListByOrg 根据 orgID 获取容器列表
func (client *DBClient) ListContainerByOrg(orgID string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("dice_org = ?", orgID)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListContainerByProject 根据 projectID 获取容器列表
func (client *DBClient) ListContainerByProject(projectID, orgID string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("dice_project = ?", projectID)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Where("dice_org = ?", orgID).
		Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListContainerByApplication 根据 appID 获取容器列表
func (client *DBClient) ListContainerByApplication(appID, orgID string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("dice_application = ?", appID)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Where("dice_org = ?", orgID).
		Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListContainerByRuntime 根据 runtimeID 获取容器列表
func (client *DBClient) ListContainerByRuntime(runtimeID, orgID string, running bool) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("dice_runtime = ?", runtimeID)
	if running {
		db = db.Where("status in (?)", allRunningStatus)
	}
	if err := db.Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Where("dice_org = ?", orgID).
		Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

// ListRunningByService 根据serviceName 获取运行中实例列表
func (client *DBClient) ListRunningContainersByService(runtimeID, serviceName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("dice_runtime = ?", runtimeID).
		Where("dice_service = ?", serviceName).
		Where("status in (?)", allRunningStatus).
		Not("dice_project", []string{"", types.UnknownType}).
		Not("dice_application", []string{"", types.UnknownType}).
		Where("dice_component = ''").
		Where("dice_addon = ''").
		Where("dice_addon_name = ''").
		Order("created_at desc").Find(&containers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return containers, nil
}

// ListEdasContainerByParam 根据 params 过滤出 EDAS 容器列表
func (client *DBClient) ListEdasContainerByParam(params *apistructs.EdasContainerListRequest) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("status in (?)", allRunningStatus)
	if len(params.EdasAppIDs) == 0 {
		db = db.Where("edas_app_id != ''")
	} else {
		db = db.Where("edas_app_id in (?)", params.EdasAppIDs)
	}
	if params.ProjectID != 0 {
		db = db.Where("dice_project = ?", params.ProjectID)
	}
	if params.AppID != 0 {
		db = db.Where("dice_application = ?", params.AppID)
	}
	if params.RuntimeID != 0 {
		db = db.Where("dice_runtime = ?", params.RuntimeID)
	}
	if params.Workspace != "" {
		db = db.Where("dice_workspace = ?", params.Workspace)
	}
	if params.Service != "" {
		db = db.Where("dice_service = ?", params.Service)
	}
	if err := db.Order("started_at desc").Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// ListClusterRunningContainers 根据clusterName 获取集群运行中实例列表
func (client *DBClient) ListClusterRunningContainers(clusterName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).
		Where("dice_service is not null and dice_service != '' and dice_service != 'unknown'").
		Where("status in (?)", allRunningStatus).
		Not("dice_project", []string{"", types.UnknownType}).
		Not("dice_application", []string{"", types.UnknownType}).
		Where("dice_component = ''").
		Where("dice_addon = ''").
		Where("dice_addon_name = ''").
		Order("started_at desc").Find(&containers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
	}

	return containers, nil
}

// ListStoppedByService 根据serviceName 获取已停止实例列表
func (client *DBClient) ListStoppedContainersByService(runtimeID, serviceName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("dice_runtime = ?", runtimeID).
		Where("dice_service = ?", serviceName).
		Where("status in (?)", allStoppedStatus).
		Not("dice_project", []string{"", types.UnknownType}).
		Not("dice_application", []string{"", types.UnknownType}).
		Where("dice_component = ''").
		Where("dice_addon = ''").
		Where("dice_addon_name = ''").
		Order("created_at desc").Find(&containers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return containers, nil
}

// ListAllContainersByService 根据serviceName 获取所有实例列表(包含运行中、已停止)
func (client *DBClient) ListAllContainersByService(runtimeID, serviceName string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("dice_runtime = ?", runtimeID).
		Where("dice_service = ?", serviceName).
		Not("dice_project", []string{"", types.UnknownType}).
		Not("dice_application", []string{"", types.UnknownType}).
		Where("dice_component = ''").
		Where("dice_addon = ''").
		Where("dice_addon_name = ''").
		Order("created_at desc").Find(&containers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return containers, nil
}

// ListRunningContainersByHost 根据 cluster & hostIP 获取运行中容器列表
func (client *DBClient) ListRunningContainersByHost(clusterName, hostIP string) ([]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).
		Where("host_private_ip_addr = ?", hostIP).
		Where("status in (?)", allRunningStatus).
		Find(&containers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
	}

	return containers, nil
}

// ListRunningAddonsByCluster 根据 clusterName 获取集群中正在运行的 addon containers
func (client *DBClient) ListRunningAddonsByCluster(clusterName string) ([]model.Container, error) {
	var containers []model.Container

	if err := client.Where("cluster = ?", clusterName).
		Where("dice_addon is not null").
		Where("dice_addon <> ''").
		Where("status in (?)", allRunningStatus).Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// GetRunningAddonByClusterAndInstanceID 根据 clusterName & addon instanceID 获取运行中的addon containers
func (client *DBClient) GetRunningAddonByClusterAndInstanceID(clusterName, instanceID string) (*[]model.Container, error) {
	var containers []model.Container
	if err := client.Where("cluster = ?", clusterName).
		Where("dice_addon = ?", instanceID).
		Where("status in (?)", allRunningStatus).Find(&containers).Error; err != nil {
		return nil, err
	}
	return &containers, nil
}

// GetContainerByTaskIDOrContainerID 根据 taskID 或 containerID 获取实例列表
// 	1. scheduler事件暂时只有taskID(marathon集群唯一, k8s集群为podID)
//  2. telegraf推送事件taskID & containerID都不为空
func (client *DBClient) GetContainerByTaskIDOrContainerID(cluster, taskID, containerID string) ([]model.Container, error) {
	var containers []model.Container
	db := client.Where("cluster = ?", cluster)
	if containerID == "" {
		db = db.Where("task_id = ?", taskID)
	} else {
		db = db.Where("(container_id = ? or task_id = ? )", containerID, taskID)
	}
	if err := db.Find(&containers).Error; err != nil {
		return nil, err
	}

	return containers, nil
}

// TODO 以下皆须重构
// UpdateContainerByPrimaryKeyID 根据主键进行更新
func (client *DBClient) UpdateContainerByPrimaryKeyID(ctx context.Context, c *types.CmContainer) error {
	if c == nil {
		return errors.Errorf("invalid params: container is null")
	}

	if c.ModelHeader.ID <= 0 {
		return errors.Errorf("invalid params: primary key <= 0")
	}

	return client.Model(&types.CmContainer{}).Updates(c).Error
}

// UpdateContainerByTaskIDOrContainerID 根据 taskID 或者 contianerID 来更新 container 信息
func (client *DBClient) UpdateContainerByTaskIDOrContainerID(ctx context.Context, c *types.CmContainer) error {
	cluster := c.Cluster
	taskID := c.TaskID
	containerID := c.ID

	if len(cluster) == 0 {
		return errors.Errorf("missing cluster params")
	}

	if len(taskID) == 0 && len(containerID) == 0 {
		return errors.Errorf("missing taskID and containerID params")
	}

	database := client.Model(&types.CmContainer{})
	if len(taskID) != 0 && len(containerID) != 0 {
		database = database.Where("cluster = ? AND (task_id = ? or container_id = ?)", cluster, taskID, containerID)
	} else if len(taskID) == 0 && len(containerID) != 0 {
		database = database.Where("cluster = ? AND container_id = ?", cluster, containerID)
	} else {
		database = database.Where("cluster = ? AND task_id = ?", cluster, taskID)
	}

	// 批量更新 container
	return database.Updates(c).Error
}

// InsertContainer 新增 container 信息
func (client *DBClient) InsertContainer(ctx context.Context, c *types.CmContainer) error {
	return client.Save(c).Error
}

// DeleteContainer 删除 container 信息
func (client *DBClient) DeleteContainer(ctx context.Context, c *types.CmContainer) error {
	var err error
	err = client.Where("cluster = ? AND dice_project = ? AND dice_runtime = ? AND dice_service = ? AND container_id = ?",
		c.Cluster, c.DiceProject, c.DiceRuntime, c.DiceService, c.ID).Delete(types.CmContainer{}).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	return err
}

// QueryContainer 获取单个容器信息
func (client *DBClient) QueryContainer(ctx context.Context, cluster string, id string) (*types.CmContainer, error) {
	var container types.CmContainer
	var err error

	if cluster == "" {
		return nil, errors.Errorf("invalid params: cluster is null")
	}

	length := len(id)
	if length < 12 {
		return nil, errors.Errorf("invalid params: container_id's length < 12")
	}

	if err = client.Where("cluster = ? AND substring(container_id, 1, ?) = ?", cluster, length, id).First(&container).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New(types.NotFound)
		}
		return nil, err
	}

	return &container, nil
}

// QueryContainerByTaskIDOrContainerID 根据 taskID 或者 containerID 查询 container 信息
func (client *DBClient) QueryContainerByTaskIDOrContainerID(ctx context.Context, cluster string, taskID string,
	containerID string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(cluster) == 0 {
		return nil, errors.Errorf("missing cluster params")
	}

	if len(taskID) == 0 && len(containerID) == 0 {
		return nil, errors.Errorf("missing taskID and containerID params")
	}

	database := client.Model(&types.CmContainer{})
	if len(taskID) != 0 && len(containerID) != 0 {
		database = database.Where("cluster = ? AND (task_id = ? or container_id = ?)", cluster, taskID, containerID)
	} else if len(taskID) == 0 && len(containerID) != 0 {
		database = database.Where("cluster = ? AND container_id = ?", cluster, containerID)
	} else {
		database = database.Where("cluster = ? AND task_id = ?", cluster, taskID)
	}

	if err := database.Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

func (client *DBClient) AllRunningProjectsContainers(ctx context.Context) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if err := client.Where("status in (?)", allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllProjectsContainersByCluster 获取整个 cluster 所有通过 dice 创建的 containers
func (client *DBClient) AllProjectsContainersByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("missing cluster params")
	}

	if err := client.Where("cluster = ?", cluster).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningProjectsContainersByCluster 获取整个 cluster 所有通过 dice 创建且正在运行的 containers
func (client *DBClient) AllRunningProjectsContainersByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("missing cluster params")
	}

	if err := client.Where("cluster = ? AND status in (?)", cluster, allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByCluster 获取整个 cluster 所有的 containers
func (client *DBClient) AllContainersByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("missing cluster params")
	}

	if err := client.Where("cluster = ?", cluster).Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByCluster 获取整个 cluster 所有正在运行的 containers
func (client *DBClient) AllRunningContainersByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("missing cluster params")
	}

	if err := client.Where("cluster = ? AND status in (?)", cluster, allRunningStatus).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByHost 获取指定 host 下所有 containers
func (client *DBClient) AllContainersByHost(ctx context.Context, cluster string, host []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" || len(host) == 0 {
		return nil, errors.Errorf("missing params: cluster = %s, host = %v", cluster, host)
	}

	if err := client.Where("cluster = ? AND host_private_ip_addr in (?)", cluster, host).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByHost 获取指定 host 下所有运行中的 containers
func (client *DBClient) AllRunningContainersByHost(ctx context.Context, cluster string, host []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" || len(host) == 0 {
		return nil, errors.Errorf("missing params: cluster = %s, host = %v", cluster, host)
	}

	if err := client.Where("cluster = ? AND host_private_ip_addr in (?) AND status in (?)", cluster, host, allRunningStatus).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// DeleteAllContainersByHost 删除指定 host 下所有 containers
func (client *DBClient) DeleteAllContainersByHost(ctx context.Context, cluster, host string) error {
	err := client.Where("cluster = ? AND host_private_ip_addr = ?", cluster, host).Delete(types.CmContainer{}).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return err
}

// AllContainersByOrg 获取指定 org 下所有 containers
func (client *DBClient) AllContainersByOrg(ctx context.Context, org string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if org == "" {
		return nil, errors.Errorf("missing org params")
	}

	if err := client.Where("dice_org = ?", org).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByOrg 获取指定 org 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByOrg(ctx context.Context, org string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if org == "" {
		return nil, errors.Errorf("missing org params")
	}

	if err := client.Where("dice_org = ? AND status in (?)", org, allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByProject 获取指定 project 下所有 containers
func (client *DBClient) AllContainersByProject(ctx context.Context, project []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(project) == 0 {
		return nil, errors.Errorf("missing project params")
	}

	if err := client.Where("dice_project in (?)", project).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByProject 获取指定 project 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByProject(ctx context.Context, project []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(project) == 0 {
		return nil, errors.Errorf("missing project params")
	}

	if err := client.Where("dice_project in (?) AND status in (?)", project, allRunningStatus).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByApplication 获取指定 appliaction 下所有 containers
func (client *DBClient) AllContainersByApplication(ctx context.Context, app []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(app) == 0 {
		return nil, errors.Errorf("missing application params")
	}

	if err := client.Where("dice_application in (?)", app).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByApplication 获取指定 appliaction 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByApplication(ctx context.Context, app []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(app) == 0 {
		return nil, errors.Errorf("missing application params")
	}

	if err := client.Where("dice_application in (?) AND status in (?)", app, allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByRuntime 获取指定 runtime 下所有 containers
func (client *DBClient) AllContainersByRuntime(ctx context.Context, runtime []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(runtime) == 0 {
		return nil, errors.Errorf("missing runtime params")
	}

	if err := client.Where("dice_runtime in (?)", runtime).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByRuntime 获取指定 runtime 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByRuntime(ctx context.Context, runtime []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if len(runtime) == 0 {
		return nil, errors.Errorf("missing runtime params")
	}

	if err := client.Where("dice_runtime in (?) AND status in (?)", runtime, allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByService 获取指定 service 下所有 containers
func (client *DBClient) AllContainersByService(ctx context.Context, runtime string, service []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if runtime == "" || len(service) == 0 {
		return nil, errors.Errorf("invalid params: runtime = %s, service = %v", runtime, service)
	}

	database := client.Where("dice_runtime = ? AND dice_service in (?)", runtime, service).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''")

	// 倒序查最新数据
	if err := database.Order("id desc").Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByService 获取指定 service 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByService(ctx context.Context, runtime string, service []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if runtime == "" || len(service) == 0 {
		return nil, errors.Errorf("invalid params: runtime = %s, service = %v", runtime, service)
	}

	if err := client.Where("dice_runtime = ? AND dice_service in (?) AND status in (?)", runtime, service, allRunningStatus).
		Where("dice_project is not null AND dice_project <> '' AND dice_project <> ?", types.UnknownType).
		Where("dice_application is not null AND dice_application <> '' AND dice_application <> ?", types.UnknownType).
		Where("dice_component = '' AND dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllComponentsByCluster 获取集群所有的 dice components containers
func (client *DBClient) AllComponentsByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("invalid param: cluster is nil")
	}

	if err := client.Where("cluster = ? AND dice_component is not null AND dice_component <> ''", cluster).
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningComponentsByCluster 获取集群所有正在运行的 dice components containers
func (client *DBClient) AllRunningComponentsByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("invalid param: cluster is nil")
	}

	if err := client.Where("cluster = ? AND status in (?)", cluster, allRunningStatus).
		Where("dice_component is not null AND dice_component <> ''").
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByComponent 获取指定 component 下所有的containers
func (client *DBClient) AllContainersByComponent(ctx context.Context, cluster string, component []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" || len(component) == 0 {
		return nil, errors.Errorf("invalid param: cluster = %s, component = %v", cluster, component)
	}

	if err := client.Where("cluster = ? AND dice_component in (?)", cluster, component).
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningContainersByComponent 获取指定 component 下所有正在运行的 containers
func (client *DBClient) AllRunningContainersByComponent(ctx context.Context, cluster string, component []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" || len(component) == 0 {
		return nil, errors.Errorf("invalid param: cluster = %s, component = %v", cluster, component)
	}

	if err := client.Where("cluster = ? AND dice_component in (?) AND status in (?)", cluster, component, allRunningStatus).
		Where("dice_addon = '' AND dice_addon_name = ''").
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllAddonsByCluster 获取集群所有的 dice addons containers
func (client *DBClient) AllAddonsByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("invalid param: cluster is nil")
	}

	if err := client.Where("cluster = ? AND dice_addon is not null AND dice_addon <> ''", cluster).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllRunningAddonsByCluster 获取集群所有运行中的 dice addons containers
func (client *DBClient) AllRunningAddonsByCluster(ctx context.Context, cluster string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" {
		return nil, errors.Errorf("invalid param: cluster is nil")
	}

	if err := client.Where("cluster = ? AND dice_addon is not null AND dice_addon <> '' AND status in (?)", cluster, allRunningStatus).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// AllContainersByAddon 获取指定 addon 下所有的containers
func (client *DBClient) AllContainersByAddon(ctx context.Context, cluster string, addon []string) ([]types.CmContainer, error) {
	var containers []types.CmContainer

	if cluster == "" || len(addon) == 0 {
		return nil, errors.Errorf("invalid param: cluster = %s, addon = %v", cluster, addon)
	}

	if err := client.Where("cluster = ? AND dice_addon in (?)", cluster, addon).
		Find(&containers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return instancesDeduplicateAndRound(containers), nil
}

// DeleteStoppedContainersByPeriod 删除指定时间前，且已经停止的 containers
func (client *DBClient) DeleteStoppedContainersByPeriod(ctx context.Context, period time.Duration) error {
	var err error
	var cs []types.CmContainer

	cstZone := time.FixedZone("CST", 8*3600) // 东八区
	now := time.Now()
	deadline := now.Add(-period).In(cstZone).Format("2006-01-02 15:04:05")

	logrus.Infof("delete stopped containers, deadline: %s", deadline)

	if err = client.Where("updated_at < ? AND status in (?)", deadline, allStoppedStatus).Find(&cs).Error; err != nil {
		return err
	}
	logrus.Infof("Going to delete containers: %v", cs)

	if err = client.Where("updated_at < ? AND status in (?)", deadline, allStoppedStatus).Delete(types.CmContainer{}).Error; err != nil {
		return err
	}

	return nil
}

func instancesDeduplicateAndRound(instances []types.CmContainer) []types.CmContainer {
	instanceMap := make(map[string]*types.CmContainer, len(instances))

	// 实例列表去重，并兼容 EDAS 容器没有 taskID 的问题
	for i, data := range instances {
		if len(data.TaskID) == 0 && len(data.ID) == 0 {
			continue
		}

		var key string
		if len(data.TaskID) != 0 {
			key = data.TaskID
		} else {
			key = data.ID
		}

		v, ok := instanceMap[key]
		if !ok {
			instanceMap[key] = &instances[i]
		} else {
			if types.ContainerStatusIndex[types.InstanceStatus(data.Status)] >
				types.ContainerStatusIndex[types.InstanceStatus(v.Status)] {
				instanceMap[key].Status = data.Status
			}

			if len(v.ID) == 0 && len(data.ID) != 0 {
				instanceMap[key].ID = data.ID
			}
		}
	}

	var instanceSlice []types.CmContainer
	for k, instance := range instanceMap {
		if len(k) != 0 {
			instance.CPU = Round(instance.CPU, 2)
			instanceSlice = append(instanceSlice, *instance)
		}
	}

	return instanceSlice
}

// GetAccumulateResource 根据resource（项目、应用、runtime）类型，获取指定集群对应的数量
func (client *DBClient) GetAccumulateResource(cluster, resource string) (uint64, error) {
	if cluster == "" {
		return 0, errors.Errorf("invalid param: cluster is nil")
	}

	if resource == "" {
		return 0, errors.Errorf("invalid param: resource is nil")
	}

	var count uint64
	if err := client.Model(&model.Container{}).
		Where("cluster = ?", cluster).
		Where(fmt.Sprintf("%s is not null and %s != ''", resource, resource)).
		Where("status in (?)", allRunningStatus).
		Select(fmt.Sprintf("count(distinct(%s))", resource)).Count(&count).
		Error; err != nil {
		return 0, err
	}

	return count, nil
}
