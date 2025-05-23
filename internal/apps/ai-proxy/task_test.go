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

package ai_proxy

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_filesystem"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	cloudstorage "github.com/erda-project/erda/internal/pkg/cloud-storage"
	"github.com/erda-project/erda/internal/pkg/cloud-storage/types"
)

type mockCloudStorage struct {
	cloudstorage.Interface
}

type mockDao struct {
	dao.DAO
}

func TestCleanupFile(t *testing.T) {
	now := time.Now()
	oldTime := now.Add(-8 * 24 * time.Hour)

	tests := []struct {
		name           string
		objects        []types.FileObject
		files          []*mcp_filesystem.McpFile
		expectedError  error
		expectedDelete int
	}{
		{
			name: "正常清理场景",
			objects: []types.FileObject{
				{
					Key:          "test1",
					Size:         1000,
					LastModified: &oldTime,
				},
				{
					Key:          "test2",
					Size:         2000,
					LastModified: &now,
				},
			},
			files: []*mcp_filesystem.McpFile{
				{
					ObjectKey: "test1",
					FileSize:  1000,
					CreatedAt: oldTime,
					Keep:      "N",
				},
				{
					ObjectKey: "test2",
					FileSize:  2000,
					CreatedAt: now,
					Keep:      "Y",
				},
			},
			expectedError:  nil,
			expectedDelete: 1,
		},
		{
			name: "孤儿文件清理",
			objects: []types.FileObject{
				{
					Key:          "orphan1",
					Size:         1000,
					LastModified: &oldTime,
				},
			},
			files:          []*mcp_filesystem.McpFile{},
			expectedError:  nil,
			expectedDelete: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				CloudStorage: &mockCloudStorage{},
				Dao:          &mockDao{},
				Config: &config.Config{
					MaxFileSize: 5000,
				},
			}

			// 模拟云存储接口
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(p.CloudStorage, "ListObjects",
				func(_ cloudstorage.Interface, _ context.Context) ([]types.FileObject, error) {
					return tt.objects, nil
				})

			patches.ApplyMethod(p.CloudStorage, "DeleteObjects",
				func(_ cloudstorage.Interface, _ context.Context, _ []string) error {
					return nil
				})

			// 模拟数据库接口
			mockClient := &mcp_filesystem.DBClient{}
			patches.ApplyMethod(p.Dao, "McpFilesystemClient",
				func(_ dao.DAO) *mcp_filesystem.DBClient {
					return mockClient
				})

			patches.ApplyMethod(mockClient, "GetFileByObjectKey",
				func(_ *mcp_filesystem.DBClient, key string) (*mcp_filesystem.McpFile, error) {
					for _, file := range tt.files {
						if file.ObjectKey == key {
							return file, nil
						}
					}
					return nil, gorm.ErrRecordNotFound
				})

			patches.ApplyMethod(mockClient, "ListMcpFiles",
				func(_ *mcp_filesystem.DBClient) ([]*mcp_filesystem.McpFile, error) {
					return tt.files, nil
				})

			patches.ApplyMethod(mockClient, "DeleteFileByKey",
				func(_ *mcp_filesystem.DBClient, _ string) error {
					return nil
				})

			err := p.CleanupFile(context.Background())
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
