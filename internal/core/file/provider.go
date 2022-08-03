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
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/safe"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/etcd"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda-proto-go/core/file/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	EtcdKeyOfKmsCmk           string        `file:"etcd_key_of_kms_cmk" env:"FILE_ETCD_KEY_OF_KMS_CMK" default:"/dice/cmdb/files/kms/key"`
	cleanExpiredFilesInterval time.Duration `file:"clean_expired_files_interval" env:"FILE_CLEAN_EXPIRED_FILES_INTERVAL" default:"5m"`

	// 文件上传限制大小，默认 300MB
	FileMaxUploadSize datasize.ByteSize `file:"file_max_upload_size" env:"FILE_MAX_UPLOAD_SIZE" default:"300MB"`
	// the size of the file parts stored in memory, the default value 32M refer to https://github.com/golang/go/blob/5c489514bc5e61ad9b5b07bd7d8ec65d66a0512a/src/net/http/request.go
	FileMaxMemorySize datasize.ByteSize `file:"file_max_memory_size" env:"FILE_MAX_MEMORY_SIZE" default:"32MB"`

	// disable file download permission validate temporarily for multi-domain
	DisableFileDownloadPermissionValidate bool `file:"disable_file_download_permission_validate" env:"DISABLE_FILE_DOWNLOAD_PERMISSION_VALIDATE" default:"false"`

	// fs
	// 修改该值的话，注意同步修改 dice.yml 中 '<%$.Storage.MountPoint%>/dice/cmdb/files:/files:rw' 容器内挂载点的值
	StorageMountPointInContainer string `file:"storage_mount_point_in_container" env:"STORAGE_MOUNT_POINT_IN_CONTAINER" default:"/files"`

	// oss
	OSS OssConfig `file:"oss"`

	// If we allow uploaded file types that can carry active content
	FileTypeCarryActiveContentAllowed bool `file:"file_type_carry_active_content_allowed" env:"FILETYPE_CARRY_ACTIVE_CONTENT_ALLOWED" default:"false"`
	// File types can carry active content, separated by comma, can add more types like jsp
	FileTypesCanCarryActiveContent []string `file:"file_types_can_carry_active_content" env:"FILETYPES_CAN_CARRY_ACTIVE_CONTENT" default:"html,js,xml,htm"`
}

type OssConfig struct {
	Endpoint     string `file:"endpoint" env:"OSS_ENDPOINT"`
	AccessID     string `file:"access_id" env:"OSS_ACCESS_ID"`
	AccessSecret string `file:"access_secret" env:"OSS_ACCESS_SECRET"`
	Bucket       string `file:"bucket" env:"OSS_BUCKET"`
	PathPrefix   string `file:"path_prefix" env:"OSS_PATH_PREFIX" default:"/dice/cmdb/files"`
}

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
	// clean expired files
	safe.Go(func() {
		ticker := time.NewTicker(p.Cfg.cleanExpiredFilesInterval)
		for range ticker.C {
			_ = p.cleanExpiredFiles()
		}
	})
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.file.FileService" || ctx.Type() == pb.FileServiceServerType() || ctx.Type() == pb.FileServiceHandlerType():
		return p.fileService
	}
	return p
}

func (p *provider) cleanExpiredFiles(_expiredAt ...time.Time) error {
	// 获取过期时间
	expiredAt := time.Unix(time.Now().Unix(), 0)
	if len(_expiredAt) > 0 {
		expiredAt = _expiredAt[0]
	}

	// 获取过期文件列表
	files, err := p.db.ListExpiredFiles(expiredAt)
	if err != nil {
		logrus.Errorf("[alert] failed to list expired files, expiredBefore: %s, err: %v", expiredAt.Format(time.RFC3339), err)
		return apierrors.ErrCleanExpiredFile.InternalError(err)
	}

	// 遍历删除文件
	for _, file := range files {
		if err := p.fileService.DeleteFile(file); err != nil {
			logrus.Errorf("[alert] failed to clean expired file, fileUUID: %s, err: %v", file.UUID, err)
			continue
		}
	}

	return nil
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
