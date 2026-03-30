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

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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
}

func newTestService(t *testing.T, cfg Config) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
CREATE TABLE ai_proxy_event (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	event VARCHAR(191) NOT NULL,
	detail VARCHAR(255) NOT NULL DEFAULT ''
);`).Error)

	return &Service{
		Config:      cfg,
		EventClient: &eventmodel.DBClient{DB: db},
	}
}
