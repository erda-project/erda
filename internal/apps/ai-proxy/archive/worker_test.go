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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
)

func requireEventDay(t *testing.T, detail, want string) {
	t.Helper()
	require.Equal(t, want, archiveDayFromDetail(detail))
}

func requireArchivePayload(t *testing.T, detail string) archiveEventDetail {
	t.Helper()
	var payload archiveEventDetail
	require.NoError(t, json.Unmarshal([]byte(detail), &payload))
	return payload
}

func requireDBDeletedPayload(t *testing.T, detail string) archiveDBDeletedEventDetail {
	t.Helper()
	var payload archiveDBDeletedEventDetail
	require.NoError(t, json.Unmarshal([]byte(detail), &payload))
	return payload
}

func TestArchiveDay_DoesNotRecordSuccessBeforeDeleteCompletes(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a"})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	oldDeleteBatch := archiveDeleteBatch
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
		archiveDeleteBatch = oldDeleteBatch
	})

	archivePutObject = func(_ *Service, _ string, reader io.Reader) error {
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	archiveObjectExists = func(_ *Service, _ context.Context, _ string) (bool, error) {
		return true, nil
	}
	archiveDeleteBatch = func(_ *Service, _ context.Context, _ time.Time) (int64, error) {
		return 0, errors.New("delete failed")
	}

	err := svc.archiveDay(ctx, day)
	require.ErrorContains(t, err, "delete failed")

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 4, total)
	require.Len(t, list, 4)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayFailed, list[1].Event)
	require.Equal(t, EventArchiveDayUploaded, list[2].Event)
	require.Equal(t, EventArchiveDayStart, list[3].Event)
	detail := requireArchivePayload(t, list[1].Detail)
	require.Equal(t, "2026-03-29", detail.Day)
	require.Equal(t, "delete failed", detail.Error)

	detail = requireArchivePayload(t, list[2].Detail)
	require.Equal(t, "2026-03-29", detail.Day)
}

func TestArchiveDay_PanicIsConvertedToFailedEvent(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a"})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	oldPutObject := archivePutObject
	t.Cleanup(func() {
		archivePutObject = oldPutObject
	})

	archivePutObject = func(_ *Service, _ string, _ io.Reader) error {
		panic("missing oss endpoint")
	}

	err := svc.archiveDay(ctx, day)
	require.ErrorContains(t, err, "panic: missing oss endpoint")

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, list, 3)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayFailed, list[1].Event)
	require.Equal(t, EventArchiveDayStart, list[2].Event)
	detail := requireArchivePayload(t, list[1].Detail)
	require.Equal(t, "2026-03-29", detail.Day)
	require.Equal(t, "panic: missing oss endpoint", detail.Error)
}

func TestArchiveDay_MissingOSSConfigReturnsMeaningfulError(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a"})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	err := svc.archiveDay(ctx, day)
	require.ErrorContains(t, err, "missing OSS archive config")
	require.ErrorContains(t, err, "AI_PROXY_AUDIT_ARCHIVE_OSS_ENDPOINT")
	require.ErrorContains(t, err, "AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_ID")
	require.ErrorContains(t, err, "AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_SECRET")
	require.ErrorContains(t, err, "AI_PROXY_AUDIT_ARCHIVE_OSS_BUCKET")

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, list, 3)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayFailed, list[1].Event)
	require.Equal(t, EventArchiveDayStart, list[2].Event)
	detail := requireArchivePayload(t, list[1].Detail)
	require.Equal(t, "2026-03-29", detail.Day)
	require.Contains(t, detail.Error, "missing OSS archive config")
	require.Contains(t, detail.Error, "AI_PROXY_AUDIT_ARCHIVE_OSS_ENDPOINT")
	require.Contains(t, detail.Error, "AI_PROXY_AUDIT_ARCHIVE_OSS_BUCKET")
}

func TestArchiveDay_SuccessWritesDayOnlyDetail(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a"})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
	})

	archivePutObject = func(_ *Service, _ string, reader io.Reader) error {
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	archiveObjectExists = func(_ *Service, _ context.Context, _ string) (bool, error) {
		return true, nil
	}

	require.NoError(t, svc.archiveDay(ctx, day))

	total, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 5, total)
	require.Len(t, list, 5)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	requireEventDay(t, list[0].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
	requireEventDay(t, list[1].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayDBDeleted, list[2].Event)
	require.Equal(t, EventArchiveDayUploaded, list[3].Event)

	dbDeleted := requireDBDeletedPayload(t, list[2].Detail)
	require.Equal(t, "2026-03-29", dbDeleted.Day)
	require.EqualValues(t, 1, dbDeleted.DeletedRowCount)

	detail := requireArchivePayload(t, list[3].Detail)
	require.Equal(t, "2026-03-29", detail.Day)
	require.EqualValues(t, 1, detail.RowCount)
	require.Positive(t, detail.RawSizeBytes)
	require.Positive(t, detail.CompressedSizeBytes)
	require.Len(t, detail.Parts, 1)
	require.Equal(t, 1, detail.Parts[0].Index)
	require.Equal(t, "ai-proxy/cluster-a/audit/archive/2026/03/audit-2026-03-29.csv.gz", detail.Parts[0].ObjectKey)
	require.EqualValues(t, 1, detail.Parts[0].RowCount)
	require.Positive(t, detail.Parts[0].RawSizeBytes)
	require.Positive(t, detail.Parts[0].CompressedSizeBytes)
}

func TestTick_ContinuesDeletingUploadedDayWithoutReExport(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a", RetentionDays: 180})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	_, err := svc.EventClient.Create(ctx, EventArchiveStart, "true")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayUploaded, `{"day":"2026-03-29","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50,"parts":[{"index":1,"object_key":"ai-proxy/cluster-a/audit/archive/2026/03/audit-2026-03-29.csv.gz","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50}]}`)
	require.NoError(t, err)
	require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, "00000000-0000-0000-0000-000000000001", day.Add(time.Hour), day.Add(time.Hour), []byte("{}")).Error)

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
	})

	putCalls := 0
	archivePutObject = func(_ *Service, _ string, _ io.Reader) error {
		putCalls++
		return nil
	}
	archiveObjectExists = func(_ *Service, _ context.Context, _ string) (bool, error) {
		return true, nil
	}

	require.NoError(t, svc.tick(ctx))
	require.Zero(t, putCalls)

	hasRows, err := svc.AuditClient.HasRowsInRange(ctx, day, day.Add(24*time.Hour))
	require.NoError(t, err)
	require.False(t, hasRows)

	success, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDaySuccess)
	require.NoError(t, err)
	require.NotNil(t, success)
	requireEventDay(t, success.Detail, "2026-03-29")

	dbDeleted, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayDBDeleted)
	require.NoError(t, err)
	require.NotNil(t, dbDeleted)
	dbDeletedDetail := requireDBDeletedPayload(t, dbDeleted.Detail)
	require.Equal(t, "2026-03-29", dbDeletedDetail.Day)
	require.EqualValues(t, 1, dbDeletedDetail.DeletedRowCount)
}

func TestTick_MarksUploadedDaySuccessWhenRowsAlreadyDeleted(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, BatchSize: 2, Name: "cluster-a", RetentionDays: 180})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, EventArchiveStart, "true")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayUploaded, `{"day":"2026-03-29","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50,"parts":[{"index":1,"object_key":"ai-proxy/cluster-a/audit/archive/2026/03/audit-2026-03-29.csv.gz","row_count":1,"raw_size_bytes":100,"compressed_size_bytes":50}]}`)
	require.NoError(t, err)

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
	})

	putCalls := 0
	archivePutObject = func(_ *Service, _ string, _ io.Reader) error {
		putCalls++
		return nil
	}
	archiveObjectExists = func(_ *Service, _ context.Context, _ string) (bool, error) {
		return true, nil
	}

	require.NoError(t, svc.tick(ctx))
	require.Zero(t, putCalls)

	success, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDaySuccess)
	require.NoError(t, err)
	require.NotNil(t, success)
	requireEventDay(t, success.Detail, "2026-03-29")

	dbDeleted, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayDBDeleted)
	require.NoError(t, err)
	require.NotNil(t, dbDeleted)
	dbDeletedDetail := requireDBDeletedPayload(t, dbDeleted.Detail)
	require.Equal(t, "2026-03-29", dbDeletedDetail.Day)
	require.EqualValues(t, 0, dbDeletedDetail.DeletedRowCount)
}

func TestArchiveDay_SplitsLargeArchiveWithinSingleQueryBatch(t *testing.T) {
	svc := newTestService(t, Config{
		Enable:                     true,
		BatchSize:                  1000,
		MaxCompressedFileSizeBytes: 1,
		Name:                       "cluster-a",
	})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	for i := 1; i <= 300; i++ {
		require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, fmt.Sprintf("00000000-0000-0000-0000-%012d", i), day.Add(time.Duration(i)*time.Minute), day.Add(time.Duration(i)*time.Minute), []byte(fmt.Sprintf(`{"payload":%q}`, strings.Repeat("x", 256)))).Error)
	}

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
	})

	var uploadedKeys []string
	archivePutObject = func(_ *Service, objectKey string, reader io.Reader) error {
		uploadedKeys = append(uploadedKeys, objectKey)
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	archiveObjectExists = func(_ *Service, _ context.Context, objectKey string) (bool, error) {
		for _, key := range uploadedKeys {
			if key == objectKey {
				return true, nil
			}
		}
		return false, nil
	}

	require.NoError(t, svc.archiveDay(ctx, day))

	require.Greater(t, len(uploadedKeys), 1)
	for i, key := range uploadedKeys {
		require.Equal(t, fmt.Sprintf("ai-proxy/cluster-a/audit/archive/2026/03/audit-2026-03-29_%d.csv.gz", i+1), key)
	}

	_, list, err := svc.ListEvents(ctx, ListRequest{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
	requireEventDay(t, list[1].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayDBDeleted, list[2].Event)
	require.Equal(t, EventArchiveDayUploaded, list[3].Event)

	dbDeleted := requireDBDeletedPayload(t, list[2].Detail)
	require.Equal(t, "2026-03-29", dbDeleted.Day)
	require.EqualValues(t, 300, dbDeleted.DeletedRowCount)

	detail := requireArchivePayload(t, list[3].Detail)
	require.Len(t, detail.Parts, len(uploadedKeys))
	require.EqualValues(t, 300, detail.RowCount)
	for i, part := range detail.Parts {
		require.Equal(t, uploadedKeys[i], part.ObjectKey)
	}
}

func TestArchiveDay_SplitsWhenThresholdHitOnLastBatchBoundary(t *testing.T) {
	// BatchSize == data count: the split check fires on the very last record of the only batch.
	// With nextList empty, we still finalize the part as multipart (suffix _1).
	svc := newTestService(t, Config{
		Enable:                     true,
		BatchSize:                  100,
		MaxCompressedFileSizeBytes: 1, // always exceeded
		Name:                       "cluster-a",
	})
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	for i := 1; i <= 100; i++ {
		require.NoError(t, svc.AuditClient.DB.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, fmt.Sprintf("00000000-0000-0000-0000-%012d", i), day.Add(time.Duration(i)*time.Minute), day.Add(time.Duration(i)*time.Minute), []byte(`{}`)).Error)
	}

	oldPutObject := archivePutObject
	oldObjectExists := archiveObjectExists
	t.Cleanup(func() {
		archivePutObject = oldPutObject
		archiveObjectExists = oldObjectExists
	})

	var uploadedKeys []string
	archivePutObject = func(_ *Service, objectKey string, reader io.Reader) error {
		uploadedKeys = append(uploadedKeys, objectKey)
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	archiveObjectExists = func(_ *Service, _ context.Context, objectKey string) (bool, error) {
		for _, key := range uploadedKeys {
			if key == objectKey {
				return true, nil
			}
		}
		return false, nil
	}

	require.NoError(t, svc.archiveDay(ctx, day))

	require.Len(t, uploadedKeys, 1)
	require.Equal(t, "ai-proxy/cluster-a/audit/archive/2026/03/audit-2026-03-29_1.csv.gz", uploadedKeys[0])

	_, list, err := svc.ListEvents(ctx, ListRequest{PageNum: 1, PageSize: 5, Day: "2026-03-29"})
	require.NoError(t, err)
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
	requireEventDay(t, list[1].Detail, "2026-03-29")
	require.Equal(t, EventArchiveDayDBDeleted, list[2].Event)
	require.Equal(t, EventArchiveDayUploaded, list[3].Event)
	dbDeleted := requireDBDeletedPayload(t, list[2].Detail)
	require.Equal(t, "2026-03-29", dbDeleted.Day)
	require.EqualValues(t, 100, dbDeleted.DeletedRowCount)
	detail := requireArchivePayload(t, list[3].Detail)
	require.EqualValues(t, 100, detail.RowCount)
	require.Len(t, detail.Parts, 1)
}

func TestMarkInterruptedIfNeeded_WritesEventsWhenDayStartHasNoEnd(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, Name: "cluster-a"})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, EventArchiveDayStart, "2026-03-29")
	require.NoError(t, err)

	require.NoError(t, svc.markInterruptedIfNeeded(ctx))

	interrupted, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayInterrupted)
	require.NoError(t, err)
	require.NotNil(t, interrupted)
	requireEventDay(t, interrupted.Detail, "2026-03-29")

	end, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayEnd)
	require.NoError(t, err)
	require.NotNil(t, end)
	requireEventDay(t, end.Detail, "2026-03-29")
}

func TestMarkInterruptedIfNeeded_NoOpWhenDayEndIsLatest(t *testing.T) {
	svc := newTestService(t, Config{Enable: true, Name: "cluster-a"})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, EventArchiveDayStart, "2026-03-29")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayEnd, "2026-03-29")
	require.NoError(t, err)

	require.NoError(t, svc.markInterruptedIfNeeded(ctx))

	interrupted, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayInterrupted)
	require.NoError(t, err)
	require.Nil(t, interrupted)
}

func TestRun_NonLeaderDoesNotMarkInterrupted(t *testing.T) {
	svc := newTestService(t, Config{
		Enable:       true,
		LoopInterval: 10 * time.Millisecond,
		Name:         "cluster-a",
	})
	ctx := context.Background()

	_, err := svc.EventClient.Create(ctx, eventmodel.EventArchiveLeaderHeartbeat, "0")
	require.NoError(t, err)
	_, err = svc.EventClient.Create(ctx, EventArchiveDayStart, "2026-03-29")
	require.NoError(t, err)

	consumed := false
	require.NoError(t, svc.EventClient.DB.Callback().Update().Before("gorm:update").Register("test:lose-leader-cas", func(tx *gorm.DB) {
		if consumed || tx.Statement.Table != "ai_proxy_event" {
			return
		}
		consumed = true
		err := tx.Session(&gorm.Session{NewDB: true}).Exec(`
UPDATE ai_proxy_event
SET detail = CAST(CAST(detail AS INTEGER) + 1 AS TEXT)
WHERE event = ?
`, eventmodel.EventArchiveLeaderHeartbeat).Error
		require.NoError(t, err)
	}))

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- svc.Run(runCtx)
	}()
	time.Sleep(30 * time.Millisecond)
	cancel()
	require.NoError(t, <-done)

	interrupted, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayInterrupted)
	require.NoError(t, err)
	require.Nil(t, interrupted)

	end, err := svc.EventClient.LatestByEvent(ctx, EventArchiveDayEnd)
	require.NoError(t, err)
	require.Nil(t, end)
}
