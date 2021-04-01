package model

// Certificate 证书信息模型
type Certificate struct {
	BaseModel
	OrgID    int64
	Name     string
	Android  string
	Ios      string
	Message  string
	Type     string // IOS发布证书/Android证书/消息推送证书
	Desc     string
	Creator  string
	Operator string
}

// TableName 设置模型对应数据库表名称
func (Certificate) TableName() string {
	return "dice_certificates"
}

// AppCertificate 应用引用证书信息模型
type AppCertificate struct {
	BaseModel
	AppID         int64
	CertificateID int64
	ApprovalID    int64
	Status        string
	Operator      string
	PushConfig    string
}

// TableName 设置模型对应数据库表名称
func (AppCertificate) TableName() string {
	return "dice_app_certificates"
}
