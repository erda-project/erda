package model

// OrgClusterRelation 企业集群关联关系
type OrgClusterRelation struct {
	BaseModel
	OrgID       uint64 `gorm:"unique_index:idx_org_cluster_id"`
	OrgName     string
	ClusterID   uint64 `gorm:"unique_index:idx_org_cluster_id"`
	ClusterName string
	Creator     string
}

// TableName 设置模型对应数据库表名称
func (OrgClusterRelation) TableName() string {
	return "dice_org_cluster_relation"
}
