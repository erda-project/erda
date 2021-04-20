// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

import (
	"io"
	"time"
)

// FileUploadResponse 文件上传响应
type FileUploadResponse struct {
	Header
	Data *File `json:"data"`
}

type FileType string

var (
	FileTypePicture FileType = "picture"
	FileTypeOther   FileType = "other"
)

// GetFileTypeByExt ext has dot
func GetFileTypeByExt(ext string) FileType {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".bmp", ".gif":
		return FileTypePicture
	default:
		return FileTypeOther
	}
}

// FileUploadResponseData 文件上传响应数据
type File struct {
	ID          uint64     `json:"id"`
	UUID        string     `json:"uuid"`
	DisplayName string     `json:"name"`
	ByteSize    int64      `json:"size"`
	DownloadURL string     `json:"url"`
	Type        FileType   `json:"type"`
	From        string     `json:"from"`
	Creator     string     `json:"creator"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	ExpiredAt   *time.Time `json:"expiredAt,omitempty"`
}

// FileDownloadFailResponse 文件下载失败响应
type FileDownloadFailResponse struct {
	Header
	Data File `json:"data"`
}

// FileUploadRequest 文件上传请求数据
type FileUploadRequest struct {
	FileNameWithExt string        `json:"fileNameWithExt,omitempty"`
	ByteSize        int64         `json:"byteSize,omitempty"`
	FileReader      io.ReadCloser `json:"fileReader,omitempty"`
	From            string        `json:"from,omitempty"`     // 文件来源，例如 issue / gittar mr
	IsPublic        bool          `json:"isPublic,omitempty"` // 是否可以无需登录直接下载
	Encrypt         bool          `json:"encrypt,omitempty"`  // 是否需要加密存储
	Creator         string        `json:"creator,omitempty"`
	ExpiredAt       *time.Time    `json:"expiredAt,omitempty"` // 过期时间
}

// BackupList 备份列表
type BackupList struct {
	ID          uint64     `json:"id"`
	UUID        string     `json:"uuid"`
	DisplayName string     `json:"name"`
	ByteSize    int64      `json:"size"`
	DownloadURL string     `json:"url"`
	Type        FileType   `json:"type"`
	From        string     `json:"from"`
	Username    string     `json:"creator", gorm:"username"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	ExpiredAt   *time.Time `json:"expiredAt,omitempty"`
	CommitID    string     `json:"commitId"`
	Remark      string     `json:"remark"`
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
