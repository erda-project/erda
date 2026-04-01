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

func TestTick_DryRunAdvancesFromDryRunDayNotSuccessDay(t *testing.T) {
	// simulate: real archive done up to 2023-01-01, dry-run already processed 2023-01-02.
	// tick should process 2023-01-03 (dryRunDay+1), not 2023-01-02 (successDay+1).
	svc := newTestService(t, Config{
		Enable:        true,
		DryRun:        true,
		Name:          "cluster-a",
		RetentionDays: 180,
	})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, EventArchiveStart, "true")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDryRun, "true")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDaySuccess, `{"day":"2023-01-01","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50,"parts":[{"index":1,"object_key":"ai-proxy/cluster-a/audit/archive/2023/01/audit-2023-01-01.csv.gz","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50}]}`)
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayDryRun, "2023-01-02")
	require.NoError(t, err)

	oldObjectExists := archiveObjectExists
	t.Cleanup(func() { archiveObjectExists = oldObjectExists })
	archiveObjectExists = func(_ *Service, _ context.Context, _ string) (bool, error) {
		return true, nil
	}

	require.NoError(t, svc.tick(ctx))

	dayStart, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayStart)
	require.NoError(t, err)
	require.NotNil(t, dayStart)
	require.Equal(t, "2023-01-03", dayStart.Detail)
}

func TestTick_RepairsInterruptedDayThenContinues(t *testing.T) {
	svc := newTestService(t, Config{
		Enable:        true,
		AutoStart:     true,
		DryRun:        true,
		Name:          "cluster-a",
		RetentionDays: 1,
	})
	ctx := context.Background()
	day := time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local)

	_, err := svc.EventClient.Create(ctx, eventmodel.EventArchiveLeaderHeartbeat, "0")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayStart, day.Format("2006-01-02"))
	require.NoError(t, err)
	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	require.NoError(t, svc.tick(ctx))

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      day.Format("2006-01-02"),
	})
	require.NoError(t, err)
	require.EqualValues(t, 6, total)
	require.Len(t, list, 6)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayDryRun, list[1].Event)
	requireEventDay(t, list[1].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayStart, list[2].Event)
	require.Equal(t, EventArchiveDayEnd, list[3].Event)
	requireEventDay(t, list[3].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayInterrupted, list[4].Event)
	requireEventDay(t, list[4].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayStart, list[5].Event)
}

func TestTick_MarksInterruptedBeforeRunningGuard(t *testing.T) {
	svc := newTestService(t, Config{
		Enable:        true,
		AutoStart:     true,
		DryRun:        true,
		Name:          "cluster-a",
		RetentionDays: 1,
	})
	ctx := context.Background()
	day := time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local)

	_, err := svc.EventClient.Create(ctx, eventmodel.EventArchiveLeaderHeartbeat, "0")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayStart, day.Format("2006-01-02"))
	require.NoError(t, err)

	require.NoError(t, svc.tick(ctx))

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      day.Format("2006-01-02"),
	})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, list, 3)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayInterrupted, list[1].Event)
	requireEventDay(t, list[1].Detail, day.Format("2006-01-02"))
	require.Equal(t, EventArchiveDayStart, list[2].Event)
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
