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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
)

func TestArchiveConfigNormalize(t *testing.T) {
	t.Run("disabled archive does not require name", func(t *testing.T) {
		cfg := Config{}
		require.NoError(t, cfg.Normalize())
	})

	t.Run("enabled archive requires name", func(t *testing.T) {
		cfg := Config{Enable: true}
		require.ErrorContains(t, cfg.Normalize(), "AI_PROXY_AUDIT_ARCHIVE_NAME")
	})
}

func TestBuildStatus_UsesEventOverAutoStart(t *testing.T) {
	cfg := Config{Enable: true, AutoStart: true, Name: "cluster-a"}

	status := buildStatus(cfg, &eventmodel.Event{Event: EventArchiveStart, Detail: "false"})
	require.True(t, status.Enabled)
	require.False(t, status.Started)

	status = buildStatus(cfg, &eventmodel.Event{Event: EventArchiveStart, Detail: "true"})
	require.True(t, status.Started)
}

func TestBuildStatus_FallsBackToAutoStart(t *testing.T) {
	cfg := Config{Enable: true, AutoStart: true, Name: "cluster-a"}

	status := buildStatus(cfg, nil)
	require.True(t, status.Enabled)
	require.True(t, status.Started)
	require.False(t, status.Running)
}

func TestBuildStatus_TracksRunningDayAndLatestResult(t *testing.T) {
	cfg := Config{Enable: true, Name: "cluster-a"}
	now := time.Now()

	status := buildStatus(
		cfg,
		&eventmodel.Event{Event: EventArchiveStart, Detail: "true", CreatedAt: now},
		&eventmodel.Event{Event: EventArchiveDayStart, Detail: "2026-03-01", CreatedAt: now.Add(time.Second)},
		&eventmodel.Event{Event: EventArchiveDayEnd, Detail: "2026-02-28", CreatedAt: now},
		&eventmodel.Event{Event: EventArchiveDayFailed, Detail: "2026-02-28", CreatedAt: now},
	)
	require.True(t, status.Running)
	require.Equal(t, "2026-03-01", status.CurrentDay)
	require.Equal(t, EventArchiveDayFailed, status.LastResult)

	status = buildStatus(
		cfg,
		&eventmodel.Event{Event: EventArchiveStart, Detail: "true", CreatedAt: now},
		&eventmodel.Event{Event: EventArchiveDayStart, Detail: "2026-03-01", CreatedAt: now},
		&eventmodel.Event{Event: EventArchiveDayEnd, Detail: "2026-03-01", CreatedAt: now.Add(time.Second)},
		&eventmodel.Event{Event: EventArchiveDaySuccess, Detail: "2026-03-01", CreatedAt: now.Add(time.Second)},
	)
	require.False(t, status.Running)
	require.Empty(t, status.CurrentDay)
	require.Equal(t, EventArchiveDaySuccess, status.LastResult)
}
