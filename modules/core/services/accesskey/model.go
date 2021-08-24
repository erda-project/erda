// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
