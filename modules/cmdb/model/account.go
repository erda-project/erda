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
