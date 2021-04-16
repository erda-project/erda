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

package dbclient

import (
	"crypto/md5" // #nosec G501
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/spec"
	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// ServiceGroup is the common resource based on which deployments are created
// projectId, git branch and working dev determine a runtime
type Runtime struct {
	dbengine.BaseModel
	Name              string `gorm:"not null;unique_index:idx_unique_app_id_name"`
	ApplicationID     uint64 `gorm:"not null;unique_index:idx_unique_app_id_name"`
	Workspace         string `gorm:"not null;unique_index:idx_unique_app_id_name"`
	GitBranch         string // Deprecated
	ProjectID         uint64 `gorm:"not null"` // TODO: currently equal to applicationID, fix later
	Env               string // Deprecated
	ClusterName       string
	ClusterId         uint64 // Deprecated: use clusterName
	Creator           string `gorm:"not null"`
	ScheduleName      ScheduleName
	Status            string `gorm:"column:runtime_status"`
	LegacyStatus      string `gorm:"column:status"`
	Deployed          bool
	Deleting          bool `gorm:"-"` // TODO: after legacyStatus removed, we use deleting instead
	Version           string
	Source            apistructs.RuntimeSource
	DiceVersion       string
	CPU               float64
	Mem               float64 // 单位: MB
	ConfigUpdatedDate *time.Time
	// Deprecated
	ReadableUniqueId string
	// Deprecated
	GitRepoAbbrev string
	OrgID         uint64 `gorm:"not null"`
}

const (
	LegacyStatusDeleting = "DELETING"
)

func (Runtime) TableName() string {
	return "ps_v2_project_runtimes"
}

type ScheduleName struct {
	Namespace string
	Name      string
}

func (r *Runtime) InitScheduleName(clusterType string, enabledPrjNamespace bool) {
	name := md5V(fmt.Sprintf("%d-%s-%s", r.ApplicationID, r.Workspace, r.Name))
	if enabledPrjNamespace {
		// 开启了项目级命名空间后，需要改成1个10位的哈希值id
		name = fnvV(fmt.Sprintf("%d-%s-%s", r.ApplicationID, r.Workspace, r.Name))
	}
	if clusterType == apistructs.EDAS {
		name = fmt.Sprintf("%s-%d", strings.ToLower(r.Workspace), r.ID)
	}
	r.ScheduleName = ScheduleName{
		Namespace: "services",
		Name:      name,
	}
}

// fnvV 生成10位的哈希值
func fnvV(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))[:10]
}

// md5V md5加密
func md5V(str string) string {
	h := md5.New() // #nosec G401
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func (s ScheduleName) Value() (driver.Value, error) {
	if s.Namespace == "" || s.Name == "" {
		return nil, nil
	}
	return strutil.Concat(s.Namespace, "/", s.Name), nil
}

func (s *ScheduleName) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		str = fmt.Sprintf("%v", v)
	}
	if str == "" {
		return nil
	}
	v := strutil.Split(str, "/", true)
	if len(v) != 2 {
		return errors.Errorf("scheduleName not format: %s", str)
	}
	s.Namespace = v[0]
	s.Name = v[1]
	return nil
}

func (s ScheduleName) Args() (string, string) {
	return s.Namespace, s.Name
}

type RuntimeService struct {
	dbengine.BaseModel
	RuntimeId   uint64 `gorm:"not null;unique_index:idx_runtime_id_service_name"`
	ServiceName string `gorm:"not null;unique_index:idx_runtime_id_service_name"`
	Cpu         string
	Mem         int
	Environment string `gorm:"type:text"`
	Ports       string
	Replica     int
	Status      string
	Errors      string `gorm:"type:text"`
}

// TableName runtime service 表名
func (RuntimeService) TableName() string {
	return "ps_runtime_services"
}

type RuntimeInstance struct {
	dbengine.BaseModel
	InstanceId string `gorm:"not null;unique_index:idx_instance_id"`
	RuntimeId  uint64 `gorm:"not null"`
	ServiceId  uint64 `gorm:"not null;index:idx_service_id"`
	Ip         string
	Status     string
	// stage记录容器退出的阶段, 只体现在增量事件中的退出(Killed, Failed, Finished)事件
	// 当前分的阶段为：
	// a) 容器启动阶段（健康检查超时之前退出）,"BeforeHealthCheckTimeout"
	// b) 健康检查超时阶段（被健康检查所杀）,"HealthCheckTimeout"
	// c) 后健康检查阶段（健康检查完成后退出）,"AfterHealthCheckTimeout"
	Stage string
}

// TableName runtime instance 表名
func (RuntimeInstance) TableName() string {
	return "ps_runtime_instances"
}

// RuntimeDomain indicated default and custom domain for endpoints
type RuntimeDomain struct {
	dbengine.BaseModel
	RuntimeId    uint64 `gorm:"not null"`
	Domain       string `gorm:"unique_index:unique_domain_key"`
	DomainType   string
	EndpointName string
	UseHttps     bool
}

func (RuntimeDomain) TableName() string {
	return "ps_v2_domains"
}

func (db *DBClient) CreateRuntime(runtime *Runtime) error {
	return db.Save(runtime).Error
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntime(id uint64) (*Runtime, error) {
	var runtime Runtime
	if err := db.
		Where("id = ?", id).
		Find(&runtime).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get runtime by id: %d", id)
	}
	return &runtime, nil
}

func (db *DBClient) GetRuntimeAllowNil(id uint64) (*Runtime, error) {
	var runtime Runtime
	result := db.
		Where("id = ?", id).
		Find(&runtime)
	if result.Error != nil {
		if result.RecordNotFound() {
			return nil, nil
		}
		return nil, errors.Wrapf(result.Error, "failed to get runtime by id: %d", id)
	}
	return &runtime, nil
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntimeByScheduleName(scheduleName string) (*Runtime, error) {
	runtime := &Runtime{}
	result := db.Where("schedule_name = ?", scheduleName).First(runtime)
	if result.Error != nil {
		return nil, errors.Wrapf(result.Error, "failed to get runtime by scheduleName: %v", scheduleName)
	}
	return runtime, nil
}

// if not found, return (nil, nil)
func (db *DBClient) FindRuntime(uniqueId spec.RuntimeUniqueId) (*Runtime, error) {
	var runtime Runtime
	result := db.
		Where("application_id = ? AND workspace = ? AND name = ?",
			uniqueId.ApplicationId, uniqueId.Workspace, uniqueId.Name).
		Take(&runtime)
	if result.Error != nil {
		if result.RecordNotFound() {
			return nil, nil
		}
		return nil, errors.Wrapf(result.Error, "failed to find runtime by uniqueId: %v", uniqueId)
	}
	return &runtime, nil
}

func (db *DBClient) FindRuntimesByIds(ids []uint64) ([]Runtime, error) {
	var runtimes []Runtime
	if len(ids) == 0 {
		return runtimes, nil
	}
	if err := db.
		Where("id in (?)", ids).
		Find(&runtimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find runtimes by Ids: %v", ids)
	}
	return runtimes, nil
}

func (db *DBClient) FindRuntimesByAppId(appId uint64) ([]Runtime, error) {
	var runtimes []Runtime
	if appId <= 0 {
		return runtimes, nil
	}
	if err := db.
		Where("application_id = ?", appId).
		Find(&runtimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find runtimes by appId: %v", appId)
	}
	return runtimes, nil
}

func (db *DBClient) FindRuntimeOrCreate(uniqueId spec.RuntimeUniqueId, operator string, source apistructs.RuntimeSource,
	clusterName string, clusterId uint64, gitRepoAbbrev string, projectID, orgID uint64) (*Runtime, bool, error) {
	runtime, err := db.FindRuntime(uniqueId)
	if err != nil {
		return nil, false, errors.Wrapf(err, "failed to find runtime or create by uniqueId: %v, operator: %v",
			uniqueId, operator)
	}
	created := false
	if runtime == nil {
		created = true
		runtime = &Runtime{
			ApplicationID:    uniqueId.ApplicationId,
			ProjectID:        projectID, // TODO: currently equal to applicationID, fix later
			Creator:          operator,
			Workspace:        uniqueId.Workspace,
			Env:              uniqueId.Workspace,
			Name:             uniqueId.Name,
			GitBranch:        uniqueId.Name,
			Status:           "Init",
			LegacyStatus:     "INIT",
			Source:           source,
			Deleting:         false,
			Deployed:         false,
			Version:          "1",
			DiceVersion:      "2",
			ClusterName:      clusterName,
			ClusterId:        clusterId,
			ReadableUniqueId: "dice-orchestrator",
			GitRepoAbbrev:    gitRepoAbbrev,
			Mem:              0.0,
			CPU:              0.0,
			OrgID:            orgID,
		}
		err = db.CreateRuntime(runtime)
		if err != nil {
			return nil, created, errors.Wrapf(err, "failed to find runtime or create by uniqueId: %v, operator: %v",
				uniqueId, operator)
		}
	}
	return runtime, created, nil
}

func (db *DBClient) FindDeletingRuntimes() ([]Runtime, error) {
	var runtimes []Runtime
	if err := db.
		Where("status = 'DELETING'").
		Find(&runtimes).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find deleting runtimes")
	}
	return runtimes, nil
}

// find runtimes newer than minId (id > minId)
func (db *DBClient) FindRuntimesNewerThan(minId uint64, limit int) ([]Runtime, error) {
	var runtimes []Runtime
	if err := db.
		Where("id > ?", minId).
		Limit(limit).
		Find(&runtimes).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find runtimes after: %d", minId)
	}
	return runtimes, nil
}

func (db *DBClient) UpdateRuntime(runtime *Runtime) error {
	if err := db.Save(runtime).Error; err != nil {
		return errors.Wrapf(err, "failed to update runtime, id: %v", runtime.ID)
	}
	return nil
}

func (db *DBClient) DeleteRuntime(runtimeId uint64) error {
	if err := db.
		Where("id = ?", runtimeId).
		Delete(&Runtime{}).Error; err != nil {
		return errors.Wrapf(err, "failed to delete runtime: %v", runtimeId)
	}
	return nil
}

// ListRuntimeByCluster 根据 clusterName 查找 runtime 列表
func (db *DBClient) ListRuntimeByCluster(clusterName string) ([]Runtime, error) {
	var runtimes []Runtime
	if err := db.Where("cluster_name = ?", clusterName).Find(&runtimes).Error; err != nil {
		return nil, err
	}

	return runtimes, nil
}

func (db *DBClient) CreateOrUpdateRuntimeService(service *RuntimeService, overrideStatus bool) error {
	var old RuntimeService
	result := db.
		Where("runtime_id = ? AND service_name = ?", service.RuntimeId, service.ServiceName).
		Take(&old)
	if result.Error != nil {
		if result.RecordNotFound() { // Create
			if err := db.Save(service).Error; err != nil {
				return errors.Wrap(err, "failed to CreateOrUpdateRuntimeService")
			}
			return nil
		}
		return errors.Wrap(result.Error, "failed to CreateOrUpdateRuntimeService")
	} else { // Update
		service.ID = old.ID
		service.CreatedAt = old.CreatedAt
		service.UpdatedAt = old.UpdatedAt
		service.Errors = old.Errors // TODO: should we change errors or not ?
		if !overrideStatus {
			// not override status, still use old.Status
			service.Status = old.Status
		}
		if err := db.Save(service).Error; err != nil {
			return errors.Wrap(err, "failed to CreateOrUpdateRuntimeService")
		}
	}
	return nil
}

func (db *DBClient) ClearRuntimeServiceErrors(serviceId uint64) error {
	if err := db.Model(&RuntimeService{}).
		Where("id = ?", serviceId).
		Update("errors", "").Error; err != nil {
		return errors.Wrapf(err, "failed to clear RuntimeService errors, serviceId: %v", serviceId)
	}
	return nil
}

func (db *DBClient) SetRuntimeServiceErrors(serviceId uint64, errs []apistructs.ErrorResponse) error {
	b, err := json.Marshal(errs)
	if err != nil {
		return errors.Wrapf(err, "failed to set RuntimeService errors, marshal failed, serviceId: %v, errs: %v",
			serviceId, errs)
	}
	if err := db.Model(&RuntimeService{}).
		Where("id = ?", serviceId).
		Update("errors", string(b)).Error; err != nil {
		return errors.Wrapf(err, "failed to set RuntimeService errors, serviceId: %v, errs: %v",
			serviceId, errs)
	}
	return nil
}

func (db *DBClient) GetInstanceCountByServiceId(serviceId uint64) (int, error) {
	var count int
	if err := db.Table("ps_runtime_instances").
		Where("service_id = ?", serviceId).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (db *DBClient) DeleteOldest(serviceId uint64, count int) error {
	return db.Where("service_id = ?", serviceId).
		Order("updated_at asc").
		Limit(count).
		Delete(&RuntimeInstance{}).Error
}

func (db *DBClient) GetRunningInstancesByServiceId(serviceId uint64) ([]RuntimeInstance, error) {
	var runtimeInstances []RuntimeInstance
	if err := db.Table("ps_runtime_instances").
		Where("service_id = ?", serviceId).
		Where("status in (?)", []string{
			apistructs.InstanceStatusHealthy,
			apistructs.InstanceStatusUnHealthy,
			apistructs.InstanceStatusStarting}).
		Find(&runtimeInstances).Error; err != nil {
		return nil, err
	}
	return runtimeInstances, nil
}

func (db *DBClient) CreateInstance(instance *RuntimeInstance) error {
	if err := db.Table("ps_runtime_instances").Create(&instance).Error; err != nil {
		return errors.Wrapf(err, "failed to create instance: %v", instance)
	}
	return nil
}

func (db *DBClient) UpdateInstance(instance *RuntimeInstance) error {
	if err := db.Table("ps_runtime_instances").Save(&instance).Error; err != nil {
		return errors.Wrapf(err, "failed to update instance: %v", instance)
	}
	return nil
}

func (db *DBClient) GetInstanceByTaskId(taskId string) (*RuntimeInstance, error) {
	var runtimeInstance RuntimeInstance
	result := db.Where("instance_id = ?", taskId).Last(&runtimeInstance)
	if result.Error != nil {
		if result.RecordNotFound() {
			return nil, nil
		}
		return nil, errors.Wrapf(result.Error, "failed to get instance by taskId: %v", taskId)
	}
	return &runtimeInstance, nil
}

func (db *DBClient) FindRuntimeServices(runtimeId uint64) ([]RuntimeService, error) {
	var services []RuntimeService
	if err := db.
		Where("runtime_id = ?", runtimeId).
		Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetRuntimeByProjectIDs 通过projectIDs获取对应runtime
func (db *DBClient) GetRuntimeByProjectIDs(projectIDs []uint64) (*[]Runtime, error) {
	var runtimes []Runtime
	if err := db.Where("project_id in (?)", projectIDs).Find(&runtimes).Error; err != nil {
		return nil, err
	}
	return &runtimes, nil
}

// TODO: we no need app, just redundant fields into runtime table
func ConvertRuntimeDTO(runtime *Runtime, app *apistructs.ApplicationDTO) *apistructs.RuntimeDTO {
	return &apistructs.RuntimeDTO{
		ID:              runtime.ID,
		Name:            runtime.Name,
		GitBranch:       runtime.Name,
		Workspace:       runtime.Workspace,
		ClusterName:     runtime.ClusterName,
		Status:          runtime.Status,
		ClusterId:       runtime.ClusterId,
		ApplicationID:   runtime.ApplicationID,
		ApplicationName: app.Name,
		ProjectID:       app.ProjectID,
		ProjectName:     app.ProjectName,
		OrgID:           app.OrgID,
	}
}

// CountServiceReferenceByClusterAndOrg 统计集群中service数量
func (db *DBClient) CountServiceReferenceByClusterAndOrg(clusterName, orgID string) (int, error) {
	var total int
	if err := db.Where("org_id = ?", orgID).
		Where("cluster_name = ?", clusterName).
		Model(&Runtime{}).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
