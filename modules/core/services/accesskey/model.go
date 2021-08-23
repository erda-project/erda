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

package accesskey

import (
	"time"

	"github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

// store secret key pair
type AccessKey struct {
	ID          string                         `json:"id" sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
	AccessKey   string                         `json:"accessKey" gorm:"size:24;unique;comment:'Access Key ID'"`
	SecretKey   string                         `json:"secretKey" gorm:"size:32;unique;comment:'Secret Key'"`
	Status      pb.StatusEnum_Status           `json:"status" gorm:"comment:'status of access key'"`
	SubjectType pb.SubjectTypeEnum_SubjectType `json:"subjectType" gorm:"comment:'authentication subject type. eg: organization, micro_service, system'"`
	Subject     string                         `json:"subject" gorm:"comment:'authentication subject identifier. eg: id, name or something'"`
	Description string                         `json:"description" gorm:"comment:'description'"`
	CreatedAt   time.Time                      `json:"createdAt" gorm:"comment:'created time'"`
	UpdatedAt   time.Time                      `json:"updatedAt"`
}

func (ak AccessKey) TableName() string {
	return "erda_access_key"
}

func (ak *AccessKey) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.NewV4().String())
}
