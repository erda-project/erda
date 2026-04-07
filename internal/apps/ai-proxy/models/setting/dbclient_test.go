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

package setting

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBClient_CreateOrUpdateAndGetByNamespaceKey(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, prepareSQLiteSettingTable(db))

	client := &DBClient{DB: db}
	ctx := context.Background()

	err = client.CreateOrUpdate(ctx, &Setting{
		Namespace: "blacklist_user_agent",
		Key:       "general.prompts",
		Value:     "claude;;;codex",
	})
	require.NoError(t, err)

	got, err := client.GetByNamespaceKey(ctx, "blacklist_user_agent", "general.prompts")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "claude;;;codex", got.Value)

	err = client.CreateOrUpdate(ctx, &Setting{
		Namespace: "blacklist_user_agent",
		Key:       "general.prompts",
		Value:     "cursor",
	})
	require.NoError(t, err)

	got, err = client.GetByNamespaceKey(ctx, "blacklist_user_agent", "general.prompts")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "cursor", got.Value)
}

func TestDBClient_GetByNamespaceKeys(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, prepareSQLiteSettingTable(db))

	client := &DBClient{DB: db}
	ctx := context.Background()

	require.NoError(t, client.CreateOrUpdate(ctx, &Setting{
		Namespace: "blacklist_user_agent",
		Key:       "general.headers",
		Value:     "codex",
	}))
	require.NoError(t, client.CreateOrUpdate(ctx, &Setting{
		Namespace: "blacklist_user_agent",
		Key:       "general.prompts",
		Value:     "claude",
	}))
	require.NoError(t, client.CreateOrUpdate(ctx, &Setting{
		Namespace: "audit",
		Key:       "archive.enabled",
		Value:     "true",
	}))

	got, err := client.GetByNamespaceKeys(ctx, "blacklist_user_agent", "general.headers", "general.prompts", "missing")
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "codex", got["general.headers"].Value)
	require.Equal(t, "claude", got["general.prompts"].Value)
}

func prepareSQLiteSettingTable(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE ai_proxy_setting (
	id CHAR(36) PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	deleted_at DATETIME NULL,
	namespace VARCHAR(191) NOT NULL,
	key VARCHAR(191) NOT NULL,
	value TEXT NOT NULL DEFAULT '',
	UNIQUE(namespace, key)
);`).Error
}
