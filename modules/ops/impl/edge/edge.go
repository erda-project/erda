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

package edge

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/ops/dbclient"
	"github.com/erda-project/erda/modules/ops/services/kubernetes"
	"github.com/erda-project/erda/pkg/clientgo/apis/openyurt/v1alpha1"
)

const (
	siteNodeCountFormat        = "%d/%d"
	DeploymentType             = "Deployment"
	StatefulSetType            = "StatefulSet"
	UnitedDeploymentAPIVersion = "apps.openyurt.io/v1alpha1"
	UnitedDeploymentKind       = "UnitedDeployment"
	SecretKind                 = "Secret"
	SecretApiVersion           = "v1"
	EdgeAppPrefix              = "edgeapp"
	EdgeAppDeployingStatus     = "deploying"
	EdgeAppSucceedStatus       = "succeed"
)

var (
	// 对象内存中映射关系目前只用于 ID/Name/调度器请求地址 等不可变参数
	// TODO:如果涉及其他参数，可调整为每次请求查询 / 开启定时数据内存同步 goroutine
	clusterInfos = make(map[int64]*apistructs.ClusterInfo, 0)
	orgInfos     = make(map[int64]*apistructs.OrgDTO, 0)
)

// NodePools
type NodePools = map[string]*v1alpha1.NodePool

// Edge 边缘站点操作封装
type Edge struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
	k8s *kubernetes.Kubernetes
}

// Option 定义 Edge 对象的配置选项
type Option func(*Edge)

// New 新建 Edge 实例，操作应用资源
func New(options ...Option) *Edge {
	site := &Edge{}
	for _, op := range options {
		op(site)
	}
	return site
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(e *Edge) {
		e.db = db
	}
}

// WithKubernetes 配置 k8s client
func WithKubernetes(k *kubernetes.Kubernetes) Option {
	return func(e *Edge) {
		e.k8s = k
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Edge) {
		e.bdl = bdl
	}
}

// getClusterInfo 从内存中获取 或根据 cluster id 查询 cluster info
func (e *Edge) getClusterInfo(clusterID int64) (*apistructs.ClusterInfo, error) {

	if clusterInfo, ok := clusterInfos[clusterID]; ok {
		return clusterInfo, nil
	}
	clusterInfo, err := e.bdl.GetCluster(strconv.FormatInt(clusterID, 10))
	if err != nil {
		logrus.Errorf("query cluster info failed, cluster:%d, err:%v", clusterID, err)
		return nil, fmt.Errorf("query cluster info failed, cluster:%d, err:%v", clusterID, err)
	}
	clusterInfos[clusterID] = clusterInfo
	return clusterInfo, nil
}

// getOrgInfo 从内存中获取 或根据 org id 查询 org info
func (e *Edge) getOrgInfo(orgID int64) (*apistructs.OrgDTO, error) {

	if orgInfo, ok := orgInfos[orgID]; ok {
		return orgInfo, nil
	}
	orgInfo, err := e.bdl.GetOrg(strconv.FormatInt(orgID, 10))
	if err != nil {
		return nil, fmt.Errorf("query org info failed, org:%d, err:%v", orgID, err)
	}
	orgInfos[orgID] = orgInfo
	return orgInfo, nil
}

func (e *Edge) IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// 将 edgeApp 存储结构转换为API所需结构
func (e *Edge) ConvertToEdgeAppInfo(edgeApp *dbclient.EdgeApp) (*apistructs.EdgeAppInfo, error) {
	var edgeSites []string
	var dependApp []string
	var portMaps []apistructs.PortMap
	var extraData map[string]string
	if len(edgeApp.EdgeSites) != 0 {
		if err := json.Unmarshal([]byte(edgeApp.EdgeSites), &edgeSites); err != nil {
			logrus.Errorf("fff, %+v", err)
			return nil, err
		}
	}
	if len(edgeApp.DependApp) != 0 {
		if err := json.Unmarshal([]byte(edgeApp.DependApp), &dependApp); err != nil {
			logrus.Errorf("gff, %+v", err)
			return nil, err
		}
	}
	if len(edgeApp.PortMaps) != 0 {
		if err := json.Unmarshal([]byte(edgeApp.PortMaps), &portMaps); err != nil {
			logrus.Errorf("cff, %+v", err)
			return nil, err
		}
	}
	if len(edgeApp.ExtraData) != 0 {
		if err := json.Unmarshal([]byte(edgeApp.ExtraData), &extraData); err != nil {
			logrus.Errorf("gff, %+v", err)
			return nil, err
		}
	}
	return &apistructs.EdgeAppInfo{
		ID:                  edgeApp.ID,
		OrgID:               edgeApp.OrgID,
		Name:                edgeApp.Name,
		ClusterID:           edgeApp.ClusterID,
		Type:                edgeApp.Type,
		Image:               edgeApp.Image,
		RegistryAddr:        edgeApp.RegistryAddr,
		RegistryUser:        edgeApp.RegistryUser,
		RegistryPassword:    edgeApp.RegistryPassword,
		HealthCheckType:     edgeApp.HealthCheckType,
		HealthCheckHttpPort: edgeApp.HealthCheckHttpPort,
		HealthCheckHttpPath: edgeApp.HealthCheckHttpPath,
		HealthCheckExec:     edgeApp.HealthCheckExec,
		ProductID:           edgeApp.ProductID,
		AddonName:           edgeApp.AddonName,
		AddonVersion:        edgeApp.AddonVersion,
		ConfigSetName:       edgeApp.ConfigSetName,
		Replicas:            edgeApp.Replicas,
		Description:         edgeApp.Description,
		EdgeSites:           edgeSites,
		DependApp:           dependApp,
		LimitCpu:            edgeApp.LimitCpu,
		RequestCpu:          edgeApp.RequestCpu,
		LimitMem:            edgeApp.LimitMem,
		RequestMem:          edgeApp.RequestMem,
		PortMaps:            portMaps,
		ExtraData:           extraData,
	}, nil
}
