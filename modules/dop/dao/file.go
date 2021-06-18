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

package dao

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/storage"
)

type File struct {
	dbengine.BaseModel

	UUID             string       // UUID，解决重名问题；用自增 id 在存储介质上容易冲突
	DisplayName      string       // 文件名，用于下载的时候展示
	Ext              string       // 文件后缀，带 dot
	ByteSize         int64        // 文件大小，Byte
	StorageType      storage.Type // 存储类型
	FullRelativePath string       // 文件相对路径，不包含 网盘挂载点 或 oss bucket
	From             string       // 文件来源，例如 issue / gittar mr
	Creator          string       // 文件创建者 ID
	Extra            FileExtra    // 额外信息，包括存储介质关键信息快照等
	ExpiredAt        *time.Time   // 文件超时自动删除
}

type FileExtra struct {
	OSSSnapshot OSSSnapshot `json:"ossSnapshot,omitempty"`
	IsPublic    bool        `json:"isPublic,omitempty"`

	Encrypt             bool   `json:"encrypt,omitempty"`
	KMSKeyID            string `json:"kmsKeyID,omitempty"`
	DEKCiphertextBase64 string `json:"dekCiphertextBase64,omitempty"`
}

type OSSSnapshot struct {
	OSSEndpoint string `json:"ossEndpoint,omitempty"`
	OSSBucket   string `json:"ossBucket,omitempty"`
}

func (File) TableName() string {
	return "dice_files"
}

func (ex FileExtra) Value() (driver.Value, error) {
	if b, err := json.Marshal(ex); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal Extra")
	} else {
		return string(b), nil
	}
}

func (ex *FileExtra) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for Extra")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, ex); err != nil {
		return errors.Wrapf(err, "failed to unmarshal Extra")
	}
	return nil
}

func (client *DBClient) CreateFile(file *File) error {
	return client.Create(file).Error
}

func (client *DBClient) GetFile(id uint64) (File, error) {
	var file File
	err := client.First(&file, id).Error
	return file, err
}

func (client *DBClient) GetFileByUUID(uuid string) (File, error) {
	var file File
	err := client.Where("uuid=?", uuid).First(&file).Error
	return file, err
}

func (client *DBClient) DeleteFile(id uint64) error {
	return client.DB.Where("id = ?", id).Delete(&File{}).Error
}

func (client *DBClient) ListExpiredFiles(expiredAt time.Time) ([]File, error) {
	var files []File
	err := client.DB.Where("expired_at is not null and expired_at <= ?", expiredAt).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}
