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
