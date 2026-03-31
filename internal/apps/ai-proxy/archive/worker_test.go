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
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

	archivePutObject = func(_ *Service, _ time.Time, reader io.Reader) error {
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	archiveObjectExists = func(_ *Service, _ context.Context, _ time.Time) (bool, error) {
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
	require.EqualValues(t, 3, total)
	require.Len(t, list, 3)
	require.Equal(t, EventArchiveDayEnd, list[0].Event)
	require.Equal(t, EventArchiveDayFailed, list[1].Event)
	require.Equal(t, EventArchiveDayStart, list[2].Event)
	require.Contains(t, list[1].Detail, "2026-03-29 | delete failed")
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

	archivePutObject = func(_ *Service, _ time.Time, _ io.Reader) error {
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
	require.Equal(t, EventArchiveDayFailed, list[1].Event)
	require.Equal(t, EventArchiveDayStart, list[2].Event)
	require.Contains(t, list[1].Detail, "2026-03-29 | panic: missing oss endpoint")
}
