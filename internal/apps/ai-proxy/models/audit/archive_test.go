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

package audit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

func TestArchiveQueriesIncludeSoftDeletedRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
CREATE TABLE ai_proxy_filter_audit (
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	deleted_at DATETIME NULL,
	metadata BLOB NOT NULL
);`).Error)

	client := &DBClient{DB: db}
	ctx := context.Background()
	day := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	rec := &Audit{BaseModel: common.BaseModelWithID("00000000-0000-0000-0000-000000000001")}
	rec.CreatedAt = day.Add(2 * time.Hour)
	rec.UpdatedAt = day.Add(2 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Exec(`
INSERT INTO ai_proxy_filter_audit (id, created_at, updated_at, deleted_at, metadata)
VALUES (?, ?, ?, NULL, ?)
`, rec.ID.String, rec.CreatedAt, rec.UpdatedAt, []byte("{}")).Error)
	require.NoError(t, db.WithContext(ctx).Exec(`
UPDATE ai_proxy_filter_audit SET deleted_at = ? WHERE id = ?
`, day.Add(3*time.Hour), rec.ID.String).Error)

	oldestDay, ok, err := client.OldestDayBefore(ctx, day.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, day, oldestDay)

	hasRows, err := client.HasRowsInRange(ctx, day, day.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, hasRows)

	count, err := client.CountRowsInRange(ctx, day, day.Add(24*time.Hour))
	require.NoError(t, err)
	require.EqualValues(t, 1, count)

	list, err := client.ListArchiveBatch(ctx, day, day.Add(24*time.Hour), nil, "", 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, rec.ID.String, list[0].ID.String)
	require.True(t, list[0].DeletedAt.Valid)

	deleted, err := client.DeleteArchiveBatch(ctx, day, day.Add(24*time.Hour), 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, deleted)

	var remaining int64
	require.NoError(t, db.WithContext(ctx).Unscoped().Model(&Audit{}).Where("id = ?", rec.ID.String).Count(&remaining).Error)
	require.Zero(t, remaining)
}
