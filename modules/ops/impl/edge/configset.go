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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
)

var (
	// config-企业ID-配置集名称   e.g. edgecfg-000-democofnigset
	defaultConfigSetFormat = "edgecfg-%s-%s"
)

// ListConfigSet 根据 OrgID/ClusterID 等条件获取 ConfigSet 信息
func (e *Edge) ListConfigSet(param *apistructs.EdgeConfigSetListPageRequest) (int, *[]apistructs.EdgeConfigSetInfo, error) {
	total, configSets, err := e.db.ListEdgeConfigSet(param)
	if err != nil {
		return 0, nil, err
	}

	configSetInfos := make([]apistructs.EdgeConfigSetInfo, 0, len(*configSets))

	for _, configSet := range *configSets {
		clusterInfo, err := e.getClusterInfo(configSet.ClusterID)
		if err != nil {
			return 0, nil, err
		}
		configSetInfos = append(configSetInfos, *convertToEdgeConfigSetInfo(configSet, clusterInfo.Name))
	}
	return total, &configSetInfos, nil

}

// GetConfigSet 获取单个配置集信息
func (e *Edge) GetConfigSet(configSetID int64) (*apistructs.EdgeConfigSetInfo, error) {
	var (
		configSetInfo *apistructs.EdgeConfigSetInfo
	)

	configSet, err := e.db.GetEdgeConfigSet(configSetID)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := e.getClusterInfo(configSet.ClusterID)
	if err != nil {
		return nil, err
	}

	configSetInfo = convertToEdgeConfigSetInfo(*configSet, clusterInfo.Name)

	return configSetInfo, nil
}

// CreateConfigSet 指定集群下创建 namespace 用于存放配置集
func (e *Edge) CreateConfigSet(req *apistructs.EdgeConfigSetCreateRequest) (uint64, error) {
	var (
		nsAPIVersion = "v1"
		nsKind       = "Namespace"
	)

	configSet := &dbclient.EdgeConfigSet{
		OrgID:       req.OrgID,
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: nsAPIVersion,
			Kind:       nsKind,
		},
	}

	clusterInfo, err := e.getClusterInfo(req.ClusterID)
	if err != nil {
		return 0, err
	}

	namespace.Name, err = e.getConfigSetName(configSet.OrgID, configSet.Name)
	if err != nil {
		return 0, err
	}

	if err = e.k8s.CreateNamespace(clusterInfo.Name, namespace); err != nil {
		return 0, err
	}

	if err = e.db.CreateEdgeConfigSet(configSet); err != nil {
		deleteErr := e.k8s.DeleteNamespace(clusterInfo.Name, namespace.Name)
		if deleteErr != nil {
			return 0, deleteErr
		}
		return 0, err
	}

	return configSet.ID, err

}

// DeleteConfigSet 指定集群下删除配置集
func (e *Edge) DeleteConfigSet(configSetID int64) error {
	// TODO: 事务
	configSet, err := e.db.GetEdgeConfigSet(configSetID)
	if err != nil {
		return fmt.Errorf("failed to get edge configset, configSetID: %d, err: (%v)", configSetID, err)
	}

	// 如果配置集被应用引用，则不能被删除
	edgeApps, err := e.db.GetEdgeAppByConfigset(configSet.Name, configSet.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to get releated app")
	}
	if len(*edgeApps) > 0 {
		relatedApps := make([]string, 0)
		for _, relatedApp := range *edgeApps {
			relatedApps = append(relatedApps, relatedApp.Name)
		}
		return fmt.Errorf("application %s  have related this configset", fmt.Sprint(relatedApps))
	}

	if err = e.db.DeleteEdgeConfigSet(configSetID); err != nil {
		return fmt.Errorf("failed to delete edge configset, (%v)", err)
	}

	clusterInfo, err := e.getClusterInfo(configSet.ClusterID)
	if err != nil {
		return err
	}

	if err = e.db.DeleteEdgeCfgSetItemByCfgID(configSetID); err != nil {
		return err
	}

	nsName, err := e.getConfigSetName(configSet.OrgID, configSet.Name)
	if err != nil {
		return err
	}

	err = e.k8s.DeleteNamespace(clusterInfo.Name, nsName)

	if err != nil {

		return err
	}

	return nil
}

// UpdateConfigSet 指定集群下更新配置集 （不允许更改 Name)
func (e *Edge) UpdateConfigSet(configSetID int64, req *apistructs.EdgeConfigSetUpdateRequest) error {

	configSet, err := e.db.GetEdgeConfigSet(configSetID)
	if err != nil {
		return fmt.Errorf("failed to get edge configset,configSetID: %d, (%v)", configSetID, err)
	}

	configSet.DisplayName = req.DisplayName
	configSet.Description = req.Description

	err = e.db.UpdateEdgeConfigSet(configSet)
	if err != nil {
		return err
	}
	return nil
}

// getConfigSetName 生成ConfigSet namespace 名称
func (e *Edge) getConfigSetName(orgID int64, configSetName string) (string, error) {
	return fmt.Sprintf(defaultConfigSetFormat, fmt.Sprintf("%03d", orgID), configSetName), nil
}

// convertToEdgeConfigSetInfo 转换 EdgeConfigSet 类型为数据返回 Info 类型
func convertToEdgeConfigSetInfo(configSet dbclient.EdgeConfigSet, clusterName string) *apistructs.EdgeConfigSetInfo {
	return &apistructs.EdgeConfigSetInfo{
		ID:          int64(configSet.ID),
		OrgID:       configSet.OrgID,
		Name:        configSet.Name,
		ClusterName: clusterName,
		DisplayName: configSet.DisplayName,
		ClusterID:   configSet.ClusterID,
		Description: configSet.Description,
		CreatedAt:   configSet.CreatedAt,
		UpdatedAt:   configSet.UpdatedAt,
	}
}
