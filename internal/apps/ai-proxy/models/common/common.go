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

package common

import (
	"time"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
)

type BaseModel struct {
	ID        fields.UUID      `gorm:"column:id;type:char(36)" json:"id" yaml:"id"`
	CreatedAt time.Time        `gorm:"column:created_at;type:datetime" json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time        `gorm:"column:updated_at;type:datetime" json:"updatedAt" yaml:"updatedAt"`
	DeletedAt fields.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt" yaml:"deletedAt"`
}

func BaseModelWithID(id string) BaseModel {
	return BaseModel{ID: fields.UUID{String: id, Valid: true}}
}
