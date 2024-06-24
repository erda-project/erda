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
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/file/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/internal/core/file/filetypes"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/pkg/common/pbutil"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/kms/kmscrypto"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	"github.com/erda-project/erda/pkg/storage"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *fileService) ifDispositionInline(fileType string) bool {
	sc := s.p.Cfg.Security
	if sc.FileTypeCarryActiveContentAllowed {
		return true
	}
	fileType = strutil.TrimPrefixes(fileType, ".")
	if strutil.Exist(sc.FileTypesCanCarryActiveContent, fileType) {
		return false
	}
	return true
}

func (s *fileService) headerValueDispositionInline(fileType, filename string) string {
	if s.ifDispositionInline(fileType) {
		return fmt.Sprintf("inline; filename=%s", filename)
	}
	return fmt.Sprintf("attachment; filename=%s", filename)
}

func (s *fileService) UploadFile(req filetypes.FileUploadRequest) (*pb.File, error) {
	// 校验文件大小
	if req.ByteSize > int64(s.p.Cfg.Limit.FileMaxUploadSize) {
		return nil, apierrors.ErrUploadTooLargeFile.InvalidParameter(errors.Errorf("max file size: %s", s.p.Cfg.Limit.FileMaxUploadSize.String()))
	}

	// 处理文件元信息
	ext := filepath.Ext(req.FileNameWithExt)
	fileUUID := uuid.UUID()
	destFileName := strutil.Concat(fileUUID, ext)
	handledPath, err := s.handleFilePath(destFileName)
	if err != nil {
		return nil, apierrors.ErrUploadFile.InvalidParameter(err)
	}

	// storager
	storager := s.GetStorage()
	fileReader := req.FileReader
	var DEKCiphertextBase64 string // 使用 CMK 加密后的 DEK 密文

	// 加密存储（信封加密）
	if req.Encrypt {
		fileBytes, err := io.ReadAll(req.FileReader)
		if err != nil {
			return nil, apierrors.ErrUploadFileEncrypt.InternalError(err)
		}
		// get DEK
		generateDEKResp, err := s.bdl.KMSGenerateDataKey(apistructs.KMSGenerateDataKeyRequest{
			GenerateDataKeyRequest: kmstypes.GenerateDataKeyRequest{
				KeyID: s.p.GetKMSKey(),
			},
		})
		if err != nil {
			return nil, apierrors.ErrUploadFileEncrypt.InternalError(err)
		}
		DEKCiphertextBase64 = generateDEKResp.CiphertextBase64
		DEK, err := base64.StdEncoding.DecodeString(generateDEKResp.PlaintextBase64)
		if err != nil {
			return nil, apierrors.ErrUploadFileEncrypt.InternalError(err)
		}
		// encrypt file content
		ciphertext, err := kmscrypto.AesGcmEncrypt(DEK, fileBytes, generateAesGemAdditionalData(req.From))
		if err != nil {
			return nil, apierrors.ErrUploadFileEncrypt.InternalError(err)
		}
		fileReader = io.NopCloser(bytes.NewBuffer(ciphertext))
	}

	if err := storager.Write(handledPath, fileReader); err != nil {
		return nil, apierrors.ErrUploadFile.InternalError(err)
	}

	// store to db
	file := db.File{
		UUID:             fileUUID,
		DisplayName:      req.FileNameWithExt,
		Ext:              ext,
		ByteSize:         req.ByteSize,
		StorageType:      storager.Type(),
		FullRelativePath: handledPath,
		From:             req.From,
		Creator:          req.Creator,
		ExpiredAt:        req.ExpiredAt,
	}
	file.Extra = s.handleFileExtra(file)
	file.Extra.IsPublic = req.IsPublic
	if req.Encrypt {
		file.Extra.Encrypt = req.Encrypt
		file.Extra.KMSKeyID = s.p.GetKMSKey()
		file.Extra.DEKCiphertextBase64 = DEKCiphertextBase64
	}
	if err := s.db.CreateFile(&file); err != nil {
		return nil, apierrors.ErrUploadFile.InternalError(err)
	}

	return s.convertDBFile(&file), nil
}

func (s *fileService) GetStorage(typ ...storage.Type) storage.Storager {
	var storageType storage.Type
	if len(typ) > 0 {
		storageType = typ[0]
	}

	switch storageType {
	case storage.TypeOSS:
		goto createOSS
	case storage.TypeFileSystem:
		goto createFS
	default:
		if s.p.Cfg.Storage.OSS.Endpoint != "" {
			goto createOSS
		}
		goto createFS
	}

createOSS:
	// TODO 这里统一用环境变量中的 oss 配置初始化客户端，如果 oss endpoint 或 bucket 发生过改变，之前通过 oss 上传的文件会下载不到。
	// 有两个方案：
	// 1. 数据迁移，将 oss 数据从老 bucket 移动到新 bucket
	// 2. 环境变量维护多份 oss 配置，根据 file 记录里的 endpoint 和 bucket 查询正确的 oss 配置
	return storage.NewOSS(s.p.Cfg.Storage.OSS.Endpoint, s.p.Cfg.Storage.OSS.AccessID, s.p.Cfg.Storage.OSS.AccessSecret, s.p.Cfg.Storage.OSS.Bucket, nil, nil)
createFS:
	return storage.NewFS()
}

func checkPath(path string) error {
	// 校验 path
	// 防止恶意 path 获取到任何文件
	if strutil.Contains(path, "../", "./") {
		return errors.Errorf("invalid path: %s", path)
	}
	return nil
}

func (s *fileService) handleFilePath(path string) (string, error) {

	if err := checkPath(path); err != nil {
		return "", err
	}

	// 加上指定前缀，限制文件访问路径
	switch s.GetStorage().Type() {
	case storage.TypeFileSystem:
		path = filepath.Join(s.p.Cfg.Storage.StorageMountPointInContainer, path)
	case storage.TypeOSS:
		path = filepath.Join(s.p.Cfg.Storage.OSS.PathPrefix, path)
		path = strings.TrimPrefix(path, "/")
	}

	return path, nil
}

func (s *fileService) getFileDownloadLink(uuid string) string {
	return fmt.Sprintf("%s/api/files/%s", s.p.Cfg.Link.UIPublicURL, uuid)
}

func (s *fileService) handleFileExtra(file db.File) db.FileExtra {
	var extra db.FileExtra
	if file.StorageType == storage.TypeOSS {
		extra.OSSSnapshot.OSSEndpoint = s.p.Cfg.Storage.OSS.Endpoint
		extra.OSSSnapshot.OSSBucket = s.p.Cfg.Storage.OSS.Bucket
	}
	return extra
}

func (s *fileService) convertDBFile(file *db.File) *pb.File {
	return &pb.File{
		ID:          uint64(file.ID),
		UUID:        file.UUID,
		DisplayName: file.DisplayName,
		ByteSize:    file.ByteSize,
		DownloadURL: s.getFileDownloadLink(file.UUID),
		FileType:    GetFileTypeByExt(file.Ext),
		From:        file.From,
		Creator:     file.Creator,
		CreatedAt:   pbutil.GetTimestamp(&file.CreatedAt),
		UpdatedAt:   pbutil.GetTimestamp(&file.UpdatedAt),
		ExpiredAt:   pbutil.GetTimestamp(file.ExpiredAt),
	}
}

func generateAesGemAdditionalData(fileFrom string) []byte {
	return []byte(fileFrom)
}
