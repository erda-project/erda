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

package file

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/pkg/storage"
)

func Test_fileService_headerValueDispositionInline(t *testing.T) {
	s := &fileService{
		p: &provider{
			Cfg: &config{
				Security: SecurityConfig{},
			},
		},
	}

	// allowed active content
	s.p.Cfg.Security.FileTypeCarryActiveContentAllowed = true
	assert.Equal(t, "inline; filename=a.html", s.headerValueDispositionInline(".html", "a.html"))
	assert.Equal(t, "inline; filename=a.js", s.headerValueDispositionInline(".js", "a.js"))

	// not allowed active content
	s.p.Cfg.Security.FileTypeCarryActiveContentAllowed = false
	s.p.Cfg.Security.FileTypesCanCarryActiveContent = []string{"html", "js", "xml", "htm"}
	assert.Equal(t, "attachment; filename=a.html", s.headerValueDispositionInline(".html", "a.html"))
	assert.Equal(t, "attachment; filename=a.js", s.headerValueDispositionInline(".js", "a.js"))
	assert.Equal(t, "inline; filename=a.png", s.headerValueDispositionInline(".png", "a.png"))
}

func Test_fileService_ifDispositionInline(t *testing.T) {
	s := &fileService{
		p: &provider{
			Cfg: &config{
				Security: SecurityConfig{},
			},
		},
	}

	// allowed active content
	s.p.Cfg.Security.FileTypeCarryActiveContentAllowed = true
	assert.True(t, s.ifDispositionInline(".html"))
	assert.True(t, s.ifDispositionInline(".htm"))

	// not allowed active content
	s.p.Cfg.Security.FileTypeCarryActiveContentAllowed = false
	s.p.Cfg.Security.FileTypesCanCarryActiveContent = []string{"html", "js", "xml", "htm"}
	assert.False(t, s.ifDispositionInline(".html"))
	assert.False(t, s.ifDispositionInline(".htm"))
	assert.True(t, s.ifDispositionInline(".png"))
}

func Test_fileService_GetStorage(t *testing.T) {
	s := &fileService{p: &provider{Cfg: &config{}}}

	// default is fs
	storager := s.GetStorage()
	assert.Equal(t, storage.TypeFileSystem, storager.Type())

	// set fs
	storager = s.GetStorage(storage.TypeFileSystem)
	assert.Equal(t, storage.TypeFileSystem, storager.Type())

	// set oss
	storager = s.GetStorage(storage.TypeOSS)
	assert.Equal(t, storage.TypeOSS, storager.Type())
}

func Test_checkPath(t *testing.T) {
	assert.Error(t, checkPath("./"))
	assert.Error(t, checkPath("../"))
}

func Test_fileService_getFileDownloadLink(t *testing.T) {
	s := &fileService{p: &provider{Cfg: &config{}}}
	s.p.Cfg.Link.UIPublicURL = "http://localhost:80"
	assert.Equal(t, "http://localhost:80/api/files/a.html", s.getFileDownloadLink("a.html"))
}

func Test_fileService_handleFileExtra(t *testing.T) {
	s := &fileService{p: &provider{Cfg: &config{}}}

	extra := s.handleFileExtra(db.File{StorageType: storage.TypeFileSystem})
	assert.Empty(t, extra.OSSSnapshot)

	s.p.Cfg.Storage.OSS.Endpoint = "test"
	s.p.Cfg.Storage.OSS.Bucket = "bucket"
	extra = s.handleFileExtra(db.File{StorageType: storage.TypeOSS})
	assert.Equal(t, "test", extra.OSSSnapshot.OSSEndpoint)
	assert.Equal(t, "bucket", extra.OSSSnapshot.OSSBucket)
}
