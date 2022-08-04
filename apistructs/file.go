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

package apistructs

import (
	"time"

	"github.com/erda-project/erda-proto-go/core/file/pb"
)

// BackupList 备份列表
type BackupList struct {
	ID          uint64      `json:"id"`
	UUID        string      `json:"uuid"`
	DisplayName string      `json:"name"`
	ByteSize    int64       `json:"size"`
	DownloadURL string      `json:"url"`
	Type        pb.FileType `json:"type"`
	From        string      `json:"from"`
	Username    string      `json:"creator" gorm:"username"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	ExpiredAt   *time.Time  `json:"expiredAt,omitempty"`
	CommitID    string      `json:"commitId"`
	Remark      string      `json:"remark"`
}

// 仓库备份文件信息
type RepoFiles struct {
	ID        uint64
	RepoID    int64
	Remark    string
	UUID      string     `json:"uuid"`
	CommitID  string     `json:"commitId"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// BackupListResponse 获取备份列表响应
type BackupListResponse struct {
	RepoFiles []BackupList `json:"files"`
	Total     int          `json:"total"`
}
