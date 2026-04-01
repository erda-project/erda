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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	auditmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
)

func TestServiceSetStartAndGetStatus(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, AutoStart: true, Name: "cluster-a"})
	ctx := context.Background()

	status, err := svc.GetStatus(ctx)
	require.NoError(t, err)
	require.True(t, status.Started)

	require.NoError(t, svc.SetStart(ctx, false))

	status, err = svc.GetStatus(ctx)
	require.NoError(t, err)
	require.False(t, status.Started)

	require.NoError(t, svc.SetDryRun(ctx, true))

	status, err = svc.GetStatus(ctx)
	require.NoError(t, err)
	require.True(t, status.DryRun)
}

func TestServiceListEvents(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, Name: "cluster-a"})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, EventArchiveDayStart, "2026-03-29")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDaySuccess, "2026-03-29")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveStart, "true")
	require.NoError(t, err)

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 2)
	require.Equal(t, EventArchiveDaySuccess, list[0].Event)

	total, list, err = svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Types:    []string{EventArchiveStart},
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, list, 1)
	require.Equal(t, EventArchiveStart, list[0].Event)

	total, list, err = svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Types:    []string{EventArchiveStart, EventArchiveDaySuccess},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 2)
	require.Equal(t, EventArchiveStart, list[0].Event)
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
}

func TestServiceDryRunDay_DoesNotDeleteAuditRows(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, DryRun: true, BatchSize: 2, Name: "cluster-a"})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	require.NoError(t, svc.dryRunDay(ctx, day, "would export archive CSV and upload to OSS"))

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, list, 3)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	require.Equal(t, EventArchiveDayDryRun, list[1].Event)
	require.Equal(t, EventArchiveDayStart, list[2].Event)

	hasRows, err := svc.AuditClient.HasRowsInRange(ctx, day, day.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, hasRows)
}

func newTestService(t *testing.T, cfg Config) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, prepareArchiveSQLiteEventTable(db))
	require.NoError(t, db.Exec(`
CREATE TABLE ai_proxy_filter_audit (
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	deleted_at DATETIME NULL,
	metadata BLOB NOT NULL
);`).Error)

	return &Service{
		Config:      cfg,
		EventClient: &eventmodel.DBClient{DB: db},
		AuditClient: &auditmodel.DBClient{DB: db},
	}
}

func prepareArchiveSQLiteEventTable(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE ai_proxy_event (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	event VARCHAR(191) NOT NULL,
	detail TEXT NOT NULL DEFAULT ''
);`).Error
}
