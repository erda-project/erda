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

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBClient_CreateLatestAndList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, prepareSQLiteEventTable(db))

	client := &DBClient{DB: db}
	ctx := context.Background()

	_, err = client.Create(ctx, EventArchiveStart, "true")
	require.NoError(t, err)
	time.Sleep(time.Millisecond)
	_, err = client.Create(ctx, EventArchiveDayStart, "2026-03-29")
	require.NoError(t, err)
	time.Sleep(time.Millisecond)
	_, err = client.Create(ctx, EventArchiveDayDryRun, "2026-03-29")
	require.NoError(t, err)
	time.Sleep(time.Millisecond)
	_, err = client.Create(ctx, EventArchiveDaySuccess, "2026-03-29")
	require.NoError(t, err)
	time.Sleep(time.Millisecond)
	_, err = client.Create(ctx, EventArchiveDayFailed, "2026-03-29 | missing oss endpoint")
	require.NoError(t, err)
	time.Sleep(time.Millisecond)
	_, err = client.Create(ctx, EventArchiveLeaderHeartbeat, "1")
	require.NoError(t, err)

	latest, err := client.LatestByEvent(ctx, EventArchiveDayFailed)
	require.NoError(t, err)
	require.NotNil(t, latest)
	require.Equal(t, "2026-03-29 | missing oss endpoint", latest.Detail)

	total, list, err := client.ListDayEvents(ctx, ListOptions{
		PageNum:  1,
		PageSize: 10,
		Day:      "2026-03-29",
	})
	require.NoError(t, err)
	require.EqualValues(t, 4, total)
	require.Len(t, list, 4)
	require.Equal(t, EventArchiveDayFailed, list[0].Event)
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
	require.Equal(t, EventArchiveDayDryRun, list[2].Event)
	require.Equal(t, EventArchiveDayStart, list[3].Event)

	total, list, err = client.ListDayEvents(ctx, ListOptions{
		PageNum:  1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.EqualValues(t, 5, total)
	require.Len(t, list, 5)
	require.Equal(t, EventArchiveDayFailed, list[0].Event)
	require.Equal(t, EventArchiveDaySuccess, list[1].Event)
	require.Equal(t, EventArchiveDayDryRun, list[2].Event)
	require.Equal(t, EventArchiveDayStart, list[3].Event)
	require.Equal(t, EventArchiveStart, list[4].Event)

	total, list, err = client.ListDayEvents(ctx, ListOptions{
		PageNum:  1,
		PageSize: 10,
		Types:    []string{EventArchiveStart},
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, list, 1)
	require.Equal(t, EventArchiveStart, list[0].Event)

	total, list, err = client.ListDayEvents(ctx, ListOptions{
		PageNum:  1,
		PageSize: 10,
		Types:    []string{EventArchiveStart, EventArchiveDaySuccess},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 2)
	require.Equal(t, EventArchiveDaySuccess, list[0].Event)
	require.Equal(t, EventArchiveStart, list[1].Event)
}

func prepareSQLiteEventTable(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE ai_proxy_event (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	event VARCHAR(191) NOT NULL,
	detail VARCHAR(255) NOT NULL DEFAULT ''
);`).Error
}
