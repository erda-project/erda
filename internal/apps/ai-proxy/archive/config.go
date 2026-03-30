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

package archive

import (
	"fmt"
	"strings"
	"time"

	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
)

const (
	EventArchiveStart          = eventmodel.EventArchiveStart
	EventArchiveDayStart       = eventmodel.EventArchiveDayStart
	EventArchiveDaySuccess     = eventmodel.EventArchiveDaySuccess
	EventArchiveDayFailed      = eventmodel.EventArchiveDayFailed
	EventArchiveDayInterrupted = eventmodel.EventArchiveDayInterrupted
	EventArchiveDayEnd         = eventmodel.EventArchiveDayEnd
)

type OSSConfig struct {
	Endpoint        string `file:"endpoint" env:"AI_PROXY_AUDIT_ARCHIVE_OSS_ENDPOINT"`
	AccessKeyID     string `file:"access_key_id" env:"AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_ID"`
	AccessKeySecret string `file:"access_key_secret" env:"AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_SECRET"`
	Bucket          string `file:"bucket" env:"AI_PROXY_AUDIT_ARCHIVE_OSS_BUCKET" default:"backup-ai"`
}

type Config struct {
	Enable        bool          `file:"enable" env:"AI_PROXY_AUDIT_ARCHIVE_ENABLE" default:"false"`
	AutoStart     bool          `file:"auto_start" env:"AI_PROXY_AUDIT_ARCHIVE_AUTO_START" default:"false"`
	RetentionDays int           `file:"retention_days" env:"AI_PROXY_AUDIT_ARCHIVE_RETENTION_DAYS" default:"180"`
	LoopInterval  time.Duration `file:"loop_interval" env:"AI_PROXY_AUDIT_ARCHIVE_LOOP_INTERVAL" default:"1m"`
	BatchSize     int           `file:"batch_size" env:"AI_PROXY_AUDIT_ARCHIVE_BATCH_SIZE" default:"1000"`
	Name          string        `file:"name" env:"AI_PROXY_AUDIT_ARCHIVE_NAME"`
	OSS           OSSConfig     `file:"oss"`
}

type Status struct {
	Enabled    bool
	AutoStart  bool
	Started    bool
	Running    bool
	CurrentDay string
	LastDay    string
	LastResult string
}

func (cfg *Config) Normalize() error {
	if !cfg.Enable {
		return nil
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("AI_PROXY_AUDIT_ARCHIVE_NAME is required when audit archive is enabled")
	}
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = 180
	}
	if cfg.LoopInterval <= 0 {
		cfg.LoopInterval = time.Minute
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}
	if strings.TrimSpace(cfg.OSS.Bucket) == "" {
		cfg.OSS.Bucket = "backup-ai"
	}
	return nil
}

func archiveDayStart(t time.Time) time.Time {
	tt := t.In(time.Local)
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
}

func buildStatus(cfg Config, latestStartEvent *eventmodel.Event, optionalEvents ...*eventmodel.Event) Status {
	status := Status{
		Enabled:   cfg.Enable,
		AutoStart: cfg.AutoStart,
	}
	if !cfg.Enable {
		return status
	}

	status.Started = cfg.AutoStart
	if latestStartEvent != nil && latestStartEvent.Event == EventArchiveStart {
		status.Started = latestStartEvent.Detail == "true"
	}

	var latestDayStart, latestDayEnd, latestResult *eventmodel.Event
	if len(optionalEvents) > 0 {
		latestDayStart = optionalEvents[0]
	}
	if len(optionalEvents) > 1 {
		latestDayEnd = optionalEvents[1]
	}
	if len(optionalEvents) > 2 {
		latestResult = optionalEvents[2]
	}

	if latestResult != nil {
		status.LastDay = latestResult.Detail
		status.LastResult = latestResult.Event
	}
	if latestDayStart != nil && isAfter(latestDayStart, latestDayEnd) {
		status.Running = true
		status.CurrentDay = latestDayStart.Detail
	}

	return status
}

func isAfter(left, right *eventmodel.Event) bool {
	if left == nil {
		return false
	}
	if right == nil {
		return true
	}
	if left.CreatedAt.After(right.CreatedAt) {
		return true
	}
	if left.CreatedAt.Before(right.CreatedAt) {
		return false
	}
	return left.ID > right.ID
}
