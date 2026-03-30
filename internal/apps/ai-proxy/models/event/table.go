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

package event

import "time"

const (
	EventArchiveStart          = "audit.archive.start"
	EventArchiveDayStart       = "audit.archive.day.start"
	EventArchiveDaySuccess     = "audit.archive.day.success"
	EventArchiveDayFailed      = "audit.archive.day.failed"
	EventArchiveDayInterrupted = "audit.archive.day.interrupted"
	EventArchiveDayEnd         = "audit.archive.day.end"
)

type Event struct {
	ID        uint64    `gorm:"column:id;type:bigint(20) unsigned;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	Event     string    `gorm:"column:event;type:varchar(191);index:idx_event_created_at,priority:1" json:"event"`
	Detail    string    `gorm:"column:detail;type:varchar(255)" json:"detail"`
}

func (*Event) TableName() string { return "ai_proxy_event" }

type Events []*Event
