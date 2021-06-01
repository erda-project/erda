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

package cluster

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/cmdb/services/container"
	"github.com/erda-project/erda/modules/cmdb/services/host"
	"github.com/erda-project/erda/modules/cmdb/services/ticket"
	"github.com/erda-project/erda/pkg/strutil"
)

var centralCluster = os.Getenv("DICE_CLUSTER")

// Cluster 集群操作封装
type Cluster struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
	h   *host.Host
	t   *ticket.Ticket
	con *container.Container
}

// Option 定义 Cluster 对象的配置选项
type Option func(*Cluster)

// New 新建 Cluster 实例，操作应用资源
func New(options ...Option) *Cluster {
	cluster := &Cluster{}
	for _, op := range options {
		op(cluster)
	}
	return cluster
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(c *Cluster) {
		c.db = db
	}
}

// WithHostService 配置 host service
func WithHostService(h *host.Host) Option {
	return func(c *Cluster) {
		c.h = h
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(c *Cluster) {
		c.bdl = bdl
	}
}

// WithTicketService 配置 ticket service
func WithTicketService(t *ticket.Ticket) Option {
	return func(c *Cluster) {
		c.t = t
	}
}

// WithContainerService 配置 container service
func WithContainerService(con *container.Container) Option {
	return func(c *Cluster) {
		c.con = con
	}
}

const (
	DiceProject     = "dice_project"
	DiceApplication = "dice_application"
	DiceRuntime     = "dice_runtime"
)

// CreateWithEvent 创建集群 & 发送事件
func (c *Cluster) CreateWithEvent(req *apistructs.ClusterCreateRequest) (int64, error) {
	// 判断同名集群是否已存在
	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return 0, err
	}
	if cluster != nil {
		return cluster.ID, nil
	}

	// 创建集群
	clusterID, err := c.Create(req)
	if err != nil {
		return 0, err
	}

	// 转化为事件所需格式
	clusterInfo, err := c.GetClusterByIDOrName(strconv.FormatInt(clusterID, 10))
	if err != nil {
		return 0, err
	}
	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.CreateAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCMDB,
		Content: clusterInfo,
	}
	// 发送集群创建事件
	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send cluster create event, (%v)", err)
		return 0, nil
	}

	return clusterID, nil
}

// Create 创建集群
func (c *Cluster) Create(req *apistructs.ClusterCreateRequest) (int64, error) {
	// 参数校验
	if err := c.checkCreateParam(req); err != nil {
		return 0, err
	}

	// 字段类型转换
	urls, err := json.MarshalIndent(c.parseURLs(req.Name, req.Type, req.WildcardDomain, req.URLs), "", "\t")
	if err != nil {
		return 0, err
	}
	settings, err := json.MarshalIndent(c.getDefaultSettings(req.Name, req.Type), "", "\t")
	if err != nil {
		return 0, err
	}
	c.fillConfig(req)
	config, err := json.MarshalIndent(req.Config, "", "\t")
	if err != nil {
		return 0, err
	}

	if req.SysConfig != nil {
		switch req.Type {
		case apistructs.K8S:
			if req.SysConfig.MainPlatform != nil {
				req.SchedulerConfig.MasterURL = strutil.Concat("inet://", req.WildcardDomain,
					"/insecure-kubernetes.default.svc.cluster.local")
			} else {
				req.SchedulerConfig.MasterURL = "http://insecure-kubernetes.default.svc.cluster.local"
			}
		case apistructs.DCOS:
			if req.SysConfig.MainPlatform != nil {
				req.SchedulerConfig.MasterURL = strutil.Concat("inet://", req.WildcardDomain, "/master.mesos")
			} else {
				req.SchedulerConfig.MasterURL = "http://master.mesos"
			}
		}
	}
	schedulerConfig, err := json.MarshalIndent(req.SchedulerConfig, "", "\t")
	if err != nil {
		return 0, err
	}
	sysConfig, err := json.MarshalIndent(req.SysConfig, "", "\t")
	if err != nil {
		return 0, err
	}
	opsConfig, err := json.MarshalIndent(req.OpsConfig, "", "\t")
	if err != nil {
		return 0, err
	}
	manageConfig, err := json.MarshalIndent(req.ManageConfig, "", "\t")
	if err != nil {
		return 0, err
	}

	// 集群信息入库
	cluster := &model.Cluster{
		OrgID:           req.OrgID,
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Type:            req.Type,
		CloudVendor:     req.CloudVendor,
		Logo:            req.Logo,
		WildcardDomain:  req.WildcardDomain,
		URLs:            string(urls),
		Settings:        string(settings),
		Config:          string(config),
		SchedulerConfig: string(schedulerConfig),
		OpsConfig:       string(opsConfig),
		SysConfig:       string(sysConfig),
		ManageConfig:    string(manageConfig),
	}
	if err := c.db.CreateCluster(cluster); err != nil {
		return 0, err
	}

	if req.OrgID != 0 {
		// 添加企业集群关联关系
		relation, err := c.db.GetOrgClusterRelationByOrgAndCluster(req.OrgID, cluster.ID)
		if err != nil {
			return 0, err
		}
		if relation != nil { // 若企业集群关系已存在，则返回
			return cluster.ID, nil
		}
		org, err := c.db.GetOrg(req.OrgID)
		if err != nil {
			return 0, err
		}
		relation = &model.OrgClusterRelation{
			OrgID:       uint64(req.OrgID),
			OrgName:     org.Name,
			ClusterID:   uint64(cluster.ID),
			ClusterName: req.Name,
		}
		if err := c.db.CreateOrgClusterRelation(relation); err != nil {
			return 0, err
		}
	}

	return cluster.ID, nil
}

// PushWithEvent 创建/更新集群 & 发送对应事件
func (c *Cluster) PushWithEvent(req *apistructs.ClusterCreateRequest) (int64, error) {
	// 参数校验
	if req.Name == "" {
		return 0, errors.Errorf("missing param name")
	}
	if req.SysConfig == nil {
		return 0, errors.Errorf("missing param sysConf")
	}

	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return 0, err
	}
	if cluster == nil {
		// 创建集群
		clusterID, err := c.Create(req)
		if err != nil {
			return 0, err
		}
		// 转化为事件所需格式
		clusterInfo, err := c.GetClusterByIDOrName(strconv.FormatInt(clusterID, 10))
		if err != nil {
			return 0, err
		}
		ev := &apistructs.EventCreateRequest{
			EventHeader: apistructs.EventHeader{
				Event:     bundle.ClusterEvent,
				Action:    bundle.CreateAction,
				OrgID:     "-1",
				ProjectID: "-1",
				TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
			},
			Sender:  bundle.SenderCMDB,
			Content: clusterInfo,
		}
		// 发送集群创建事件
		if err = c.bdl.CreateEvent(ev); err != nil {
			logrus.Warnf("[alert]failed to send cluster create event, (%v)", err)
			return 0, err
		}
		return clusterID, nil
	}

	// 更新集群(目前仅更新sysConf)
	sysConf, err := json.MarshalIndent(req.SysConfig, "", "\t")
	if err != nil {
		return 0, err
	}
	cluster.SysConfig = string(sysConf)
	if err = c.db.UpdateCluster(cluster); err != nil {
		return 0, err
	}

	return cluster.ID, nil
}

// UpdateWithEvent 更新集群 & 发送事件
func (c *Cluster) UpdateWithEvent(orgID int64, req *apistructs.ClusterUpdateRequest) (int64, error) {
	clusterID, err := c.Update(orgID, req)
	if err != nil {
		return 0, err
	}
	cluster, err := c.GetClusterByIDOrName(strconv.FormatInt(clusterID, 10))
	if err != nil {
		return 0, err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.UpdateAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCMDB,
		Content: cluster,
	}
	// 发送集群创建事件
	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send cluster create event, (%v)", err)
		return 0, nil
	}

	return clusterID, nil
}

// Update 更新集群
func (c *Cluster) Update(orgID int64, req *apistructs.ClusterUpdateRequest) (int64, error) {
	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return 0, err
	}
	if cluster == nil {
		return 0, errors.Errorf("not found")
	}
	logrus.Infof("before updated cluster info: %+v", cluster)
	if cluster.OrgID != orgID {
		return 0, errors.Errorf("not found")
	}

	if req.SchedulerConfig != nil {
		var oldSchedulerConfig *apistructs.ClusterSchedConfig
		if err := json.Unmarshal([]byte(cluster.SchedulerConfig), &oldSchedulerConfig); err != nil {
			return 0, err
		}
	}

	cluster.DisplayName = req.DisplayName
	if req.Type != "" {
		cluster.Type = req.Type
	}
	cluster.Logo = req.Logo
	cluster.Description = req.Description
	if req.WildcardDomain != "" {
		cluster.WildcardDomain = req.WildcardDomain
	}
	if req.URLs != nil {
		urls, err := json.MarshalIndent(c.parseURLs(req.Name, req.Type, req.WildcardDomain, req.URLs), "", "\t")
		if err != nil {
			return 0, err
		}
		cluster.URLs = string(urls)
	}
	if req.SchedulerConfig != nil {
		schedulerConfig, err := json.MarshalIndent(req.SchedulerConfig, "", "\t")
		if err != nil {
			return 0, err
		}
		cluster.SchedulerConfig = string(schedulerConfig)
	}
	if req.OpsConfig != nil {
		opsConfig, err := json.MarshalIndent(req.OpsConfig, "", "\t")
		if err != nil {
			return 0, err
		}
		cluster.OpsConfig = string(opsConfig)
	}
	if req.SysConfig != nil {
		var newSysconfig *apistructs.Sysconf
		if err := json.Unmarshal([]byte(cluster.SysConfig), &newSysconfig); err != nil {
			return 0, err
		}
		// 检验某些不变的字段是否改变；若改变，则报错
		if err := c.diffSysconfig(req.SysConfig, newSysconfig); err != nil {
			return 0, err
		}
		sysConfig, err := json.MarshalIndent(req.SysConfig, "", "\t")
		if err != nil {
			return 0, err
		}
		cluster.SysConfig = string(sysConfig)
	}
	if req.ManageConfig != nil {
		manageConfig, err := json.MarshalIndent(req.ManageConfig, "", "\t")
		if err != nil {
			return 0, err
		}
		cluster.ManageConfig = string(manageConfig)
	}

	if err := c.db.UpdateCluster(cluster); err != nil {
		return 0, nil
	}

	return cluster.ID, nil
}

// DeleteWithEvent 删除集群 & 发送事件
func (c *Cluster) DeleteWithEvent(clusterName string) error {
	cluster, err := c.db.GetClusterByName(clusterName)
	if err != nil {
		return apierrors.ErrDeleteCluster.NotFound()
	}
	logrus.Infof("deleting cluster: %+v", *cluster)

	// 删除集群
	if err := c.DeleteByName(clusterName); err != nil {
		return apierrors.ErrDeleteCluster.InternalError(err)
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.DeleteAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCMDB,
		Content: cluster,
	}
	// 发送集群删除事件
	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("[alert]failed to send cluster delete event, (%v)", err)
	}

	return nil
}

// DeleteByName 根据 clusterName 删除集群
func (c *Cluster) DeleteByName(clusterName string) error {
	// 删除企业集群关系
	if err := c.db.DeleteOrgClusterRelationByCluster(clusterName); err != nil {
		return err
	}

	// 删除集群元信息
	if err := c.db.DeleteCluster(clusterName); err != nil {
		return err
	}

	return nil
}

// GetClusterByIDOrName 根据 id/clusterName 获取集群详情
func (c *Cluster) GetClusterByIDOrName(idOrName string) (*apistructs.ClusterInfo, error) {
	var cluster *model.Cluster
	clusterID, err := strutil.Atoi64(idOrName)
	if err == nil {
		cluster, err = c.db.GetCluster(clusterID)
	} else {
		cluster, err = c.db.GetClusterByName(idOrName)
	}
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.Errorf("not found")
	}

	return c.convert(cluster), nil
}

// ListCluster 获取全部集群列表
func (c *Cluster) ListCluster() (*[]apistructs.ClusterInfo, error) {
	clusters, err := c.db.ListCluster()
	if err != nil {
		return nil, err
	}
	clusterInfos := make([]apistructs.ClusterInfo, 0, len(*clusters))
	for i := range *clusters {
		item := c.convert(&(*clusters)[i])
		// 敏感信息置空
		if item.SchedConfig != nil {
			item.SchedConfig.AuthPassword = ""
			item.SchedConfig.ClientCrt = ""
			item.SchedConfig.CACrt = ""
			item.SchedConfig.ClientKey = ""
			if item.SchedConfig.CPUSubscribeRatio == "" {
				item.SchedConfig.CPUSubscribeRatio = "1"
			}
		}
		delete(item.Settings, "nexusPassword")
		delete(item.Config, "nexusPassword")

		clusterInfos = append(clusterInfos, *item)
	}

	return &clusterInfos, nil
}

// ListClusterByOrg 根据 orgID 获取集群列表
func (c *Cluster) ListClusterByOrg(orgID int64) (*[]apistructs.ClusterInfo, error) {
	relations, err := c.db.GetOrgClusterRelationsByOrg(orgID)
	if err != nil {
		return nil, err
	}
	clusterIDs := make([]uint64, 0, len(relations))
	for _, v := range relations {
		clusterIDs = append(clusterIDs, v.ClusterID)
	}
	clusters, err := c.db.ListClusterByIDs(clusterIDs)
	if err != nil {
		return nil, err
	}
	// 设置cluster是否与企业是关联关系，而不是原创关系
	clustersInOrgs, err := c.db.ListClusterByOrg(orgID)
	if err != nil {
		return nil, err
	}
	clusterNameMap := make(map[string]string, len(*clustersInOrgs))
	if len(*clustersInOrgs) > 0 {
		for _, v := range *clustersInOrgs {
			clusterNameMap[v.Name] = ""
		}
	}

	clusterInfos := make([]apistructs.ClusterInfo, 0, len(*clusters))
	for i := range *clusters {
		item := c.convert(&(*clusters)[i])
		// 敏感信息置空
		if item.SchedConfig != nil {
			item.SchedConfig.AuthPassword = ""
			item.SchedConfig.ClientCrt = ""
			item.SchedConfig.CACrt = ""
			item.SchedConfig.ClientKey = ""
			if item.SchedConfig.CPUSubscribeRatio == "" {
				item.SchedConfig.CPUSubscribeRatio = "1"
			}
		}
		delete(item.Settings, "nexusPassword")
		delete(item.Config, "nexusPassword")
		// 设置cluster是否与企业是关联关系，而不是原创关系
		if _, ok := clusterNameMap[item.Name]; ok {
			item.IsRelation = "N"
		} else {
			item.IsRelation = "Y"
		}
		clusterInfos = append(clusterInfos, *item)
	}

	return &clusterInfos, nil
}

func (c *Cluster) ListClusterByOrgAndType(orgID int64, clusterType string) (*[]apistructs.ClusterInfo, error) {
	var clusterList []apistructs.ClusterInfo
	if orgID > 0 {
		// filter by orgID
		response, err := c.ListClusterByOrg(orgID)
		if err != nil {
			return nil, err
		}
		// filter by clusterType
		if clusterType != "" {
			for _, v := range *response {
				if v.Type == clusterType {
					clusterList = append(clusterList, v)
				}
			}
		} else {
			clusterList = *response
		}
	} else {
		clusters, err := c.db.ListClusterByOrgAndType(orgID, clusterType)
		if err != nil {
			return nil, err
		}
		clusterList = make([]apistructs.ClusterInfo, 0, len(*clusters))
		for i := range *clusters {
			item := c.convert(&(*clusters)[i])
			// 敏感信息置空
			if item.SchedConfig != nil {
				item.SchedConfig.AuthPassword = ""
				item.SchedConfig.ClientCrt = ""
				item.SchedConfig.CACrt = ""
				item.SchedConfig.ClientKey = ""
				if item.SchedConfig.CPUSubscribeRatio == "" {
					item.SchedConfig.CPUSubscribeRatio = "1"
				}
			}
			delete(item.Settings, "nexusPassword")
			delete(item.Config, "nexusPassword")

			clusterList = append(clusterList, *item)
		}
	}
	return &clusterList, nil
}

// TODO 参数校验待加强
func (c *Cluster) checkCreateParam(req *apistructs.ClusterCreateRequest) error {
	if req.Name == "" {
		return errors.Errorf("name is empty")
	}

	switch req.Type {
	case apistructs.DCOS, apistructs.EDAS, apistructs.K8S:
	default:
		return errors.Errorf("type is invalid")
	}
	return nil
}

func (c *Cluster) diffSysconfig(new, old *apistructs.Sysconf) error {
	if new.Cluster.Name != old.Cluster.Name {
		return errors.Errorf("cluster name mismatch")
	}
	if new.FPS.Host != old.FPS.Host {
		return errors.Errorf("fps host mismatch")
	}
	if new.Platform.WildcardDomain != old.Platform.WildcardDomain {
		return errors.Errorf("wildcard domain mismatch")
	}
	if new.Storage.MountPoint != old.Storage.MountPoint {
		return errors.New("nas mount point mismatch")
	}
	if new.Docker.DataRoot != old.Docker.DataRoot {
		return errors.New("docker data root mismatch")
	}
	newNodes := make(map[string]string, len(new.Nodes))
	for _, node := range new.Nodes {
		newNodes[node.IP] = node.Type
	}
	for _, node := range old.Nodes {
		if _, ok := newNodes[node.IP]; !ok {
			return errors.New("nodes info mismatch")
		}
	}
	return nil
}

// 将 cluster 存储结构转换为API所需结构
func (c *Cluster) convert(cluster *model.Cluster) *apistructs.ClusterInfo {
	var (
		urls            = make(map[string]string)
		settings        = make(map[string]interface{})
		config          = make(map[string]string)
		schedulerConfig *apistructs.ClusterSchedConfig
		opsConfig       *apistructs.OpsConfig
		manageConfig    *apistructs.ManageConfig
	)

	if cluster.URLs != "" {
		if err := json.Unmarshal([]byte(cluster.URLs), &urls); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}
	if cluster.Settings != "" {
		if err := json.Unmarshal([]byte(cluster.Settings), &settings); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)

		}
	}
	if cluster.Config != "" {
		if err := json.Unmarshal([]byte(cluster.Config), &config); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}
	if cluster.SchedulerConfig != "" {
		if err := json.Unmarshal([]byte(cluster.SchedulerConfig), &schedulerConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}
	if cluster.OpsConfig != "" {
		if err := json.Unmarshal([]byte(cluster.OpsConfig), &opsConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}

	if cluster.ManageConfig != "" {
		if err := json.Unmarshal([]byte(cluster.ManageConfig), &manageConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}

	return &apistructs.ClusterInfo{
		ID:             int(cluster.ID),
		OrgID:          int(cluster.OrgID),
		Name:           cluster.Name,
		DisplayName:    cluster.DisplayName,
		Description:    cluster.Description,
		Type:           cluster.Type,
		CloudVendor:    cluster.CloudVendor,
		Logo:           cluster.Logo,
		WildcardDomain: cluster.WildcardDomain,
		URLs:           urls,
		Settings:       settings,
		Config:         config,
		SchedConfig:    schedulerConfig,
		OpsConfig:      opsConfig,
		ManageConfig:   manageConfig,
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
	}
}

// TODO 服务端口硬编码，不妥
func (c *Cluster) parseURLs(clusterName, clusterType, wildcardDomain string, urls map[string]string) map[string]string {
	m := make(map[string]string, len(urls))
	for k, v := range urls {
		m[k] = v
	}
	if _, ok := m["colonySoldier"]; !ok {
		if clusterName == centralCluster {
			m["colonySoldier"] = strutil.Concat("http://soldier.", c.getVipSuffix(clusterType), ":9028")
			if _, ok := m["colonySoldierPublic"]; !ok {
				m["colonySoldierPublic"] = strutil.Concat("http://soldier.", wildcardDomain)
			}
		} else {
			m["colonySoldier"] = strutil.Concat("http://soldier.", wildcardDomain)
		}
	}
	if _, ok := m["nexus"]; !ok {
		m["nexus"] = strutil.Concat("http://nexus.", c.getVipSuffix(clusterType), ":8081")
	}
	if _, ok := m["registry"]; !ok {
		m["registry"] = strutil.Concat("http://registry.", c.getVipSuffix(clusterType), ":5000")
	}
	return m
}

func (c *Cluster) getDefaultSettings(clusterName, clusterType string) map[string]string {
	m := map[string]string{
		"nexusUsername":        "admin",
		"nexusPassword":        "admin123",
		"ciHostPath":           "/netdata/devops/ci",
		"registryHostPath":     "/netdata/dice/registry",
		"bpDockerBaseRegistry": strutil.Concat("registry.", c.getVipSuffix(clusterType), ":5000"),
	}
	if clusterType == apistructs.DCOS {
		m["registryAppID"] = "/devops/registry"
	}
	if clusterName != centralCluster {
		m["mainClusterName"] = centralCluster
	}
	return m
}

func (c *Cluster) fillConfig(req *apistructs.ClusterCreateRequest) {
	if len(req.Config) == 0 {
		req.Config = make(map[string]string, len(req.URLs)+len(req.Settings)+10)
	}
	req.Config["nexusUsername"] = "admin"
	req.Config["nexusPassword"] = "admin123"
	req.Config["ciHostPath"] = "/netdata/devops/ci"
	req.Config["registryHostPath"] = "/netdata/dice/registry"
	req.Config["bpDockerBaseRegistry"] = strutil.Concat("registry.", c.getVipSuffix(req.Type), ":5000")
	if req.Name != centralCluster {
		req.Config["mainClusterName"] = centralCluster
	}
	if req.Type == apistructs.DCOS {
		req.Config["registryAppID"] = "/devops/registry"
	}

	// 设置 urls
	if _, ok := req.URLs["colonySoldier"]; !ok {
		if req.Name == centralCluster {
			req.Config["colonySoldier"] = strutil.Concat("http://soldier.", req.WildcardDomain)
		} else {
			req.Config["colonySoldier"] = strutil.Concat("http://soldier.", c.getVipSuffix(req.Type), ":9028")
			if _, ok := req.URLs["colonySoldierPublic"]; !ok {
				req.Config["colonySoldierPublic"] = strutil.Concat("http://soldier.", req.WildcardDomain)
			}
		}
	}
	if _, ok := req.URLs["nexus"]; !ok {
		req.Config["nexus"] = strutil.Concat("http://nexus.", c.getVipSuffix(req.Type), ":8081")
	}
	if _, ok := req.URLs["registry"]; !ok {
		req.Config["registry"] = strutil.Concat("http://registry.", c.getVipSuffix(req.Type), ":5000")
	}
}

// TODO k8s 集群严重依赖 default namespace，不妥
func (c *Cluster) getVipSuffix(clusterType string) string {
	if clusterType == apistructs.K8S {
		return "default.svc.cluster.local"
	}
	return "marathon.l4lb.thisdcos.directory"
}

// StatisticsClusterResource 根据给定集群统计项目，应用，主机，异常主机和runtime的数量
func (c *Cluster) StatisticsClusterResource(cluster, orgName string, hostsNum uint64) (map[string]uint64, error) {
	// get project num
	projects, err := c.db.GetAccumulateResource(cluster, DiceProject)
	if err != nil {
		return nil, err
	}

	// get application num
	applications, err := c.db.GetAccumulateResource(cluster, DiceApplication)
	if err != nil {
		return nil, err
	}

	// get runtime num
	runtimes, err := c.db.GetAccumulateResource(cluster, DiceRuntime)
	if err != nil {
		return nil, err
	}

	// get hosts num
	hosts, err := c.db.GetHostsNumberByClusterAndOrg(cluster, orgName)
	if err != nil {
		return nil, err
	}

	return map[string]uint64{
		"projects":      projects,
		"applications":  applications,
		"runtimes":      runtimes,
		"hosts":         hosts,
		"abnormalHosts": hostsNum,
	}, nil
}

// ListClusterServicesList 获取指定集群的服务和addon列表
func (c *Cluster) ListClusterServices(cluster string) ([]apistructs.ServiceUsageData, error) {
	return c.con.ListClusterServices(cluster)
}

// DereferenceCluster 解除关联集群关系
func (c *Cluster) DereferenceCluster(userID string, req *apistructs.DereferenceClusterRequest) error {
	clusterInfo, err := c.db.GetClusterByName(req.Cluster)
	if err != nil {
		return err
	}
	if clusterInfo == nil {
		return errors.Errorf("不存在的集群%s", req.Cluster)
	}
	if clusterInfo.OrgID == req.OrgID {
		return errors.Errorf("非关联集群，不可解除关系")
	}
	referenceResp, err := c.bdl.FindClusterResource(req.Cluster, strconv.FormatInt(req.OrgID, 10))
	if err != nil {
		return err
	}
	if referenceResp.AddonReference > 0 || referenceResp.ServiceReference > 0 {
		return errors.Errorf("集群中存在未清理的Addon或Service，请清理后再执行.")
	}
	if err := c.db.DeleteOrgClusterRelationByClusterAndOrg(req.Cluster, req.OrgID); err != nil {
		return err
	}

	return nil
}
