package model

// ClusterEdgeSiteRelation 集群边缘站点关联关系
type ClusterEdgeSiteRelation struct {
	BaseModel
	ClusterID  int64 `gorm:"unique_index:idx_cluster_edgesite_id"`
	EdgeSiteID int64 `gorm:"unique_index:idx_cluster_edgesite_id"`
}

// TableName 设置模型对应数据库表名称
func (ClusterEdgeSiteRelation) TableName() string {
	return "cluster_edgesite_relation"
}
