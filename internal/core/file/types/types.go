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

package types

import (
	"io"
	"time"
)

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
