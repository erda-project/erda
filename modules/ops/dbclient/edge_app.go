package dbclient

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

// ListEdgeApp 分页获取边缘应用列表
func (c *DBClient) ListEdgeApp(param *apistructs.EdgeAppListPageRequest) (int, *[]EdgeApp, error) {
	var (
		total    int
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if param.OrgID < 0 {
		return 0, nil, fmt.Errorf("illegal orgin id")
	}

	db = db.Where("org_id = ?", param.OrgID).Order("id")

	if err := db.Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
		Find(&edgeApps).Error; err != nil {
		return 0, nil, err
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, &edgeApps, nil
}

// ListAllEdgeApp 获取当前企业下所有边缘应用列表
func (c *DBClient) ListAllEdgeApp(orgID int64) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 {
		return nil, fmt.Errorf("illegal orgin id")
	}

	if err := db.Where("org_id = ?", orgID).Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListAllEdgeAppByClusterID 获取当前企业指定集群下所有边缘应用列表
func (c *DBClient) ListAllEdgeAppByClusterID(orgID, clusterID int64) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 {
		return nil, fmt.Errorf("illegal orgin id")
	}

	if err := db.Where("org_id = ? and cluster_id = ?", orgID, clusterID).Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListDependsEdgeApps 获取某个应用依赖的应用列表
func (c *DBClient) ListDependsEdgeApps(orgID, clusterID int64, appName string) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 || clusterID < 0 {
		return nil, fmt.Errorf("illegal orgID or clusterID")
	}

	if err := db.Where("org_id = ? and cluster_id = ? and depend_app like ?",
		orgID, clusterID, "%"+"\""+appName+"\""+"%").Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListEdgeAppBySiteName 通过站点名获取应用列表
func (c *DBClient) ListEdgeAppBySiteName(orgID, clusterID int64, siteName string) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 || clusterID < 0 {
		return nil, fmt.Errorf("illegal orgin or cluster id")
	}

	if err := db.Where("org_id = ? and cluster_id = ? and edge_sites like ?", orgID,
		clusterID, "%"+"\""+siteName+"\""+"%").Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// GetEdgeSite 获取边缘应用详情
func (c *DBClient) GetEdgeApp(edgeAppID int64) (*EdgeApp, error) {
	var edgeApp EdgeApp
	if err := c.Where("id = ?", edgeAppID).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeSite 使用Name获取边缘应用详情
func (c *DBClient) GetEdgeAppByName(edgeName string, orgID int64) (*EdgeApp, error) {
	var edgeApp EdgeApp
	if err := c.Where("name = ? AND org_id = ?", edgeName, orgID).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeAppsBySiteName 根据站点名获取App
func (c *DBClient) GetEdgeAppsBySiteName(siteName string, clusterID int64) (*[]EdgeApp, error) {
	var edgeApp []EdgeApp
	if err := c.Where("cluster_id = ? and edge_sites like ?", clusterID, "%"+"\""+siteName+"\""+"%").Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeAppByConfigset 根据configSet获取边缘应用详情
func (c *DBClient) GetEdgeAppByConfigset(configSetName string, clusterID int64) (*[]EdgeApp, error) {
	var edgeApp []EdgeApp
	if err := c.Where("cluster_id = ? AND config_set_name = ?", clusterID, configSetName).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &edgeApp, nil
}

// CreateEdgeSite 创建边缘应用
func (c *DBClient) CreateEdgeApp(edgeApp *EdgeApp) error {
	return c.Create(edgeApp).Error
}

// UpdateEdgeSite 更新边缘应用
func (c *DBClient) UpdateEdgeApp(edgeApp *EdgeApp) error {
	return c.Save(edgeApp).Error
}

// DeleteEdgeSite 删除边缘应用
func (c *DBClient) DeleteEdgeApp(edgeAppID int64) error {
	return c.Where("id = ?", edgeAppID).Delete(&EdgeApp{}).Error
}
