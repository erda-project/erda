package model

// CurrentOrg 用户当前所属企业
type CurrentOrg struct {
	BaseModel
	UserID string
	OrgID  int64
}

// TableName 设置模型对应数据库表名称
func (CurrentOrg) TableName() string {
	return "ps_user_current_org"
}
