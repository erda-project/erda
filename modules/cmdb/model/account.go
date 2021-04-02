package model

// CloudAccount 云账号模型
type CloudAccount struct {
	BaseModel
	CloudProvider   string // 云厂商
	Name            string // 账号名
	AccessKeyID     string // KeyID, 不明文展示
	AccessKeySecret string // KeySecret, 不明文展示
	OrgID           int64  // 应用关联组织Id
}

func (CloudAccount) TableName() string {
	return "dice_cloud_accounts"
}
