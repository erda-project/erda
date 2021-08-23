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

package model

import (
	"time"
)

// store secret key pair
type AccessKey struct {
	ID          int64     `json:"id" gorm:"primary_key;comment:'Primary Key'"`
	AccessKeyID string    `json:"accessKeyId" gorm:"size:24;unique;comment:'Access Key ID'"`
	SecretKey   string    `json:"secretKey" gorm:"size:32;unique;comment:'Secret Key'"`
	IsSystem    bool      `json:"isSystem" gorm:"comment:'identify weather used for system component communication'"`
	Status      string    `json:"status" gorm:"size:16;comment:'status of access key'"`
	SubjectType string    `json:"subjectType" gorm:"comment:'authentication subject type. eg: organization, micro_service'"`
	Subject     string    `json:"subject" gorm:"comment:'authentication subject identifier. eg: id, name or something'"`
	Description string    `json:"description" gorm:"comment:'description'"`
	CreatedAt   time.Time `json:"createdAt" gorm:"comment:'created time'"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ak AccessKey) TableName() string {
	return "erda_access_key"
}
