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
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/c2h5oh/datasize"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/etcd"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda-proto-go/core/file/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type (
	config struct {
		Kms      KmsConfig      `file:"kms"`
		Cleanup  CleanupConfig  `file:"cleanup"`
		Limit    LimitConfig    `file:"limit"`
		Storage  StorageConfig  `file:"storage"`
		Security SecurityConfig `file:"security"`
		Link     LinkConfig     `file:"link"`
	}

	KmsConfig struct {
		CmkEtcdKey string `file:"cmk_etcd_key" env:"FILE_KMS_CMD_ETCD_KEY" default:"/dice/cmdb/files/kms/key"`
	}

	CleanupConfig struct {
		ExpiredFilesInterval time.Duration `file:"expired_files_interval" env:"FILE_CLEANUP_EXPIRED_FILES_INTERVAL" default:"5m"`
	}

	LimitConfig struct {
		// file upload limit size, default 300MB
		FileMaxUploadSize datasize.ByteSize `file:"file_max_upload_size" env:"FILE_MAX_UPLOAD_SIZE" default:"300MB"`
		// the size of the file parts stored in memory, the default value 32M refer to https://github.com/golang/go/blob/5c489514bc5e61ad9b5b07bd7d8ec65d66a0512a/src/net/http/request.go
		FileMaxMemorySize datasize.ByteSize `file:"file_max_memory_size" env:"FILE_MAX_MEMORY_SIZE" default:"32MB"`
	}

	StorageConfig struct {
		// fs
		// pay attention to sync the value to dice.yml '<%$.Storage.MountPoint%>/dice/cmdb/files:/files:rw' if you change this value
		StorageMountPointInContainer string `file:"storage_mount_point_in_container" env:"STORAGE_MOUNT_POINT_IN_CONTAINER" default:"/files"`

		// oss
		OSS OssConfig `file:"oss"`
	}
	OssConfig struct {
		Endpoint     string `file:"endpoint" env:"OSS_ENDPOINT"`
		AccessID     string `file:"access_id" env:"OSS_ACCESS_ID"`
		AccessSecret string `file:"access_secret" env:"OSS_ACCESS_SECRET"`
		Bucket       string `file:"bucket" env:"OSS_BUCKET"`
		PathPrefix   string `file:"path_prefix" env:"OSS_PATH_PREFIX" default:"/dice/cmdb/files"`
	}

	SecurityConfig struct {
		// disable file download permission validate temporarily for multi-domain
		DisableFileDownloadPermissionValidate bool `file:"disable_file_download_permission_validate" env:"DISABLE_FILE_DOWNLOAD_PERMISSION_VALIDATE" default:"false"`

		// If we allow uploaded file types that can carry active content
		FileTypeCarryActiveContentAllowed bool `file:"file_type_carry_active_content_allowed" env:"FILETYPE_CARRY_ACTIVE_CONTENT_ALLOWED" default:"false"`
		// File types can carry active content, separated by comma, can add more types like jsp
		FileTypesCanCarryActiveContent []string `file:"file_types_can_carry_active_content" env:"FILETYPES_CAN_CARRY_ACTIVE_CONTENT" default:"html,js,xml,htm"`
	}

	LinkConfig struct {
		UIPublicURL string `file:"ui_public_url" env:"UI_PUBLIC_URL"`
	}
)

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	MySQL      mysql.Interface
	db         *db.Client
	Etcd       etcd.Interface
	etcdClient *clientv3.Client
	bdl        *bundle.Bundle

	fileService *fileService

	kmsKey string
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.db = &db.Client{DB: p.MySQL.DB()}
	p.etcdClient = p.Etcd.Client()
	p.bdl = bundle.New(bundle.WithKMS())
	p.fileService = &fileService{
		p:   p,
		db:  p.db,
		bdl: p.bdl,
	}

	if err := p.applyKmsCmk(); err != nil {
		return fmt.Errorf("failed to apply kms cmk, err: %v", err)
	}

	if p.Register != nil {
		pb.RegisterFileServiceImp(p.Register, p.fileService, apis.Options())
		p.Register.Add(http.MethodPost, "/api/files", p.UploadFile)
		p.Register.Add(http.MethodGet, "/api/files/{uuid}", p.DownloadFile)
		p.Register.Add(http.MethodGet, "/api/files", p.DownloadFile)
		p.Register.Add(http.MethodHead, "/api/files/{uuid}", p.HeadFile)
		p.Register.Add(http.MethodDelete, "/api/files/{uuid}", p.DeleteFile)
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.asyncCleanupExpiredFiles()
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.file.FileService" || ctx.Type() == pb.FileServiceServerType() || ctx.Type() == pb.FileServiceHandlerType():
		return p.fileService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.file", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
