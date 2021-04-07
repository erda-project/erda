package edge

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
)

const (
	ScopePublic            = "public"
	ScopePublicDisplayName = "通用"
	ScopeSite              = "site"
	cmAPIVersion           = "v1"
	cmKind                 = "ConfigMap"
)

// ListConfigSetItem 根据 SiteID / Search 为条件获取配置项信息
func (e *Edge) ListConfigSetItem(param *apistructs.EdgeCfgSetItemListPageRequest) (int, *[]apistructs.EdgeCfgSetItemInfo, error) {
	total, cfgItems, err := e.db.ListEdgeConfigSetItem(param)
	if err != nil {
		return 0, nil, err
	}

	cfgSetItemInfos := make([]apistructs.EdgeCfgSetItemInfo, 0, len(*cfgItems))

	for _, cfgItem := range *cfgItems {
		if cfgItem.SiteID == 0 {
			cfgSetItemInfos = append(cfgSetItemInfos, *convertToEdgeCfgSetItemInfo(cfgItem, ScopePublic, ScopePublicDisplayName))
			continue
		}
		siteInfo, err := e.db.GetEdgeSite(cfgItem.SiteID)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to get releated edge site(%d): %v", cfgItem.SiteID, err)
		}
		cfgSetItemInfos = append(cfgSetItemInfos, *convertToEdgeCfgSetItemInfo(cfgItem, siteInfo.Name, siteInfo.DisplayName))
	}
	return total, &cfgSetItemInfos, nil
}

// GetConfigSetItem 获取单个配置项对象
func (e *Edge) GetConfigSetItem(itemID int64) (*apistructs.EdgeCfgSetItemInfo, error) {
	var (
		itemInfo *apistructs.EdgeCfgSetItemInfo
	)

	item, err := e.db.GetEdgeConfigSetItem(itemID)
	if err != nil {
		return nil, err
	}

	if item.Scope == ScopePublic {
		itemInfo = convertToEdgeCfgSetItemInfo(*item, ScopePublic, ScopePublicDisplayName)
		return itemInfo, nil
	}

	site, err := e.db.GetEdgeSite(item.SiteID)
	if err != nil {
		return nil, err
	}

	itemInfo = convertToEdgeCfgSetItemInfo(*item, site.Name, site.DisplayName)

	return itemInfo, nil
}

// CreateConfigSetItem  创建配置项
func (e *Edge) CreateConfigSetItem(req *apistructs.EdgeCfgSetItemCreateRequest) ([]uint64, error) {
	var (
		configSetItemIDs = make([]uint64, 0)
		err              error
	)

	cfgSet, err := e.db.GetEdgeConfigSet(req.ConfigSetID)
	if err != nil {
		return nil, err
	}

	nsName, err := e.getConfigSetName(cfgSet.OrgID, cfgSet.Name)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := e.getClusterInfo(cfgSet.ClusterID)
	if err != nil {
		return nil, err
	}

	if req.Scope == ScopePublic {
		err = e.createItemToConfigMap(clusterInfo.Name, nsName, ScopePublic, req.ItemKey, req.ItemValue)
		if err != nil {
			return nil, err
		}
		cfgSetItem := &dbclient.EdgeConfigSetItem{
			ConfigsetID: int64(cfgSet.ID),
			ItemKey:     req.ItemKey,
			ItemValue:   req.ItemValue,
		}

		cfgSetItem.Scope = ScopePublic

		_, res, err := e.db.ListEdgeConfigSetItem(&apistructs.EdgeCfgSetItemListPageRequest{
			ConfigSetID: int64(cfgSet.ID),
			Search:      req.ItemKey,
			SiteID:      0,
		})

		if err != nil {
			return nil, err
		}
		if len(*res) == 0 {
			err = e.db.CreateEdgeConfigSetItem(cfgSetItem)
			if err != nil {
				// 如果创建记录失败，回退已经添加至 configmap 字段
				err = e.deleteItemToConfigMap(clusterInfo.Name, nsName, ScopePublic, req.ItemKey)
				if err != nil {
					return nil, err
				}
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("item had exsited in database")
		}

		configSetItemIDs = append(configSetItemIDs, cfgSetItem.ID)
		return configSetItemIDs, nil
	}

	// TODO: ugly code
	if req.Scope == ScopeSite {
		for _, siteID := range req.SiteIDs {
			site, err := e.db.GetEdgeSite(siteID)
			if err != nil {
				return nil, err
			}
			cm, err := e.k8s.GetConfigMap(clusterInfo.Name, nsName, site.Name)

			if isResourceNotFound(err) {
				continue
			} else if err != nil {
				return nil, err
			} else {
				if _, ok := cm.Data[req.ItemKey]; ok {
					return nil, fmt.Errorf("item key %s for site %s already existed in configset %s (scope: %s), please update it", req.ItemKey, site.Name, cfgSet.Name, ScopeSite)
				}
			}
		}

		createdCm := make([]string, 0)
		for _, siteID := range req.SiteIDs {
			site, err := e.db.GetEdgeSite(siteID)
			if err != nil {
				return nil, err
			}
			err = e.createItemToConfigMap(clusterInfo.Name, nsName, site.Name, req.ItemKey, req.ItemValue)
			if err != nil {
				return nil, err
			}
			createdCm = append(createdCm, site.Name)

			cfgSetItem := &dbclient.EdgeConfigSetItem{
				ConfigsetID: int64(cfgSet.ID),
				ItemKey:     req.ItemKey,
				ItemValue:   req.ItemValue,
			}
			cfgSetItem.Scope = ScopeSite
			cfgSetItem.SiteID = int64(site.ID)

			createError := e.db.CreateEdgeConfigSetItem(cfgSetItem)
			// 创建过程存在错误进行回退，仍然可能导致错误，TODO 事务
			if createError != nil {
				backError := make([]string, 0)
				for _, cmName := range createdCm {
					err = e.deleteItemToConfigMap(clusterInfo.Name, nsName, cmName, req.ItemKey)
					if err != nil {
						backError = append(backError, err.Error())
						continue
					}
				}
				for _, cfgSetItemID := range configSetItemIDs {
					err = e.db.DeleteEdgeConfigSetItem(int64(cfgSetItemID))
					if err != nil {
						backError = append(backError, err.Error())
						continue
					}
				}
				if len(backError) != 0 {
					return nil, fmt.Errorf(fmt.Sprint(backError))
				}
				return nil, createError
			}
			configSetItemIDs = append(configSetItemIDs, cfgSetItem.ID)
		}
		return configSetItemIDs, nil
	}

	return configSetItemIDs, nil
}

// createItemToConfigMap 创建配置项信息至 ConfigMap Data
func (e *Edge) createItemToConfigMap(clusterName, namespace, cmName, itemKey, itemValue string) error {
	var (
		defaultCM = &v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       cmKind,
				APIVersion: cmAPIVersion,
			},
			Data: map[string]string{itemKey: itemValue},
		}
	)
	cm, err := e.k8s.GetConfigMap(clusterName, namespace, cmName)

	if isResourceNotFound(err) {
		defaultCM.Name = cmName
		err = e.k8s.CreateConfigMap(clusterName, namespace, defaultCM)
		if err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}

	if _, ok := cm.Data[itemKey]; ok {
		return fmt.Errorf("item key %s already existed in %s (scope: %s), please update it", itemKey, cmName, ScopePublic)
	} else {
		// 对于 configmap 存在但是数据为 0 的情况
		if cm.Data == nil {
			cm.Data = make(map[string]string, 0)
		}
		cm.Data[itemKey] = itemValue
		err = e.k8s.UpdateConfigMap(clusterName, namespace, cm)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteConfigSetItem 删除配置项信息
func (e *Edge) DeleteConfigSetItem(itemID int64) error {
	cfgSetItem, err := e.db.GetEdgeConfigSetItem(itemID)
	if err != nil {
		return err
	}
	cfgSet, err := e.db.GetEdgeConfigSet(cfgSetItem.ConfigsetID)
	if err != nil {
		return err
	}
	nsName, err := e.getConfigSetName(cfgSet.OrgID, cfgSet.Name)
	if err != nil {
		return err
	}

	clusterInfo, err := e.getClusterInfo(cfgSet.ClusterID)
	if err != nil {
		return err
	}

	if cfgSetItem.Scope == ScopePublic {
		err = e.deleteItemToConfigMap(clusterInfo.Name, nsName, ScopePublic, cfgSetItem.ItemKey)
		if err != nil {
			return err
		}
	} else if cfgSetItem.Scope == ScopeSite {
		site, err := e.db.GetEdgeSite(cfgSetItem.SiteID)
		if err != nil {
			return err
		}

		err = e.deleteItemToConfigMap(clusterInfo.Name, nsName, site.Name, cfgSetItem.ItemKey)
		if err != nil {
			return err
		}
	}

	err = e.db.DeleteEdgeConfigSetItem(itemID)
	if err != nil {
		// TODO: Update configmap which deleted.
		return err
	}

	return nil
}

// deleteItemToConfigMap 删除 Configmap 中指定 itemKey
func (e *Edge) deleteItemToConfigMap(clusterName, namespace, cmName, itemKey string) error {
	cm, err := e.k8s.GetConfigMap(clusterName, namespace, cmName)
	if isResourceNotFound(err) {
		return fmt.Errorf("configmap %s not found in cluster: %s, namespace: %s", cmName, clusterName, namespace)
	}
	if err != nil {
		return err
	}

	delete(cm.Data, itemKey)
	err = e.k8s.UpdateConfigMap(clusterName, namespace, cm)
	if err != nil {
		return err
	}
	return nil
}

// UpdateConfigSetItem 更新配置项
func (e *Edge) UpdateConfigSetItem(itemID int64, req *apistructs.EdgeCfgSetItemUpdateRequest) error {
	var cm *v1.ConfigMap

	cfgSetItem, err := e.db.GetEdgeConfigSetItem(itemID)
	if err != nil {
		return err
	}

	cfgSet, err := e.db.GetEdgeConfigSet(cfgSetItem.ConfigsetID)
	if err != nil {
		return err
	}

	nsName, err := e.getConfigSetName(cfgSet.OrgID, cfgSet.Name)
	if err != nil {
		return err
	}

	clusterInfo, err := e.getClusterInfo(cfgSet.ClusterID)
	if err != nil {
		return err
	}

	if cfgSetItem.Scope == ScopePublic {
		cm, err = e.k8s.GetConfigMap(clusterInfo.Name, nsName, ScopePublic)
		if err != nil {
			return err
		}

	} else if cfgSetItem.Scope == ScopeSite {
		site, err := e.db.GetEdgeSite(cfgSetItem.SiteID)
		if err != nil {
			return err
		}

		cm, err = e.k8s.GetConfigMap(clusterInfo.Name, nsName, site.Name)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("no scope name %s", cfgSetItem.Scope)
	}

	cm.Data[cfgSetItem.ItemKey] = req.ItemValue

	err = e.k8s.UpdateConfigMap(clusterInfo.Name, nsName, cm)
	if err != nil {
		return err
	}

	cfgSetItem.ItemValue = req.ItemValue

	err = e.db.UpdateEdgeConfigSetItem(cfgSetItem)
	if err != nil {
		// TODO: 更新回退
		return err
	}
	return nil
}

// convertToEdgeCfgSetItemInfo 转换 EdgeConfigSetItem 为返回 Info 类型
func convertToEdgeCfgSetItemInfo(item dbclient.EdgeConfigSetItem, siteName, siteDisplayName string) *apistructs.EdgeCfgSetItemInfo {
	return &apistructs.EdgeCfgSetItemInfo{
		ID:              int64(item.ID),
		ConfigSetID:     item.ConfigsetID,
		SiteID:          item.SiteID,
		SiteName:        siteName,
		SiteDisplayName: siteDisplayName,
		ItemKey:         item.ItemKey,
		ItemValue:       item.ItemValue,
		Scope:           item.Scope,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

// isResourceNotFound 判定报错是否由于资源不存在导致
func isResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "not found") {
		return true
	}
	return false
}
