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

package handler_mcp_filesystem

import (
	"context"
	"errors"
	"fmt"
	http1 "net/http"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/filesystem/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_filesystem"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	cloudstorage "github.com/erda-project/erda/internal/pkg/cloud-storage"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type MCPFileHandler struct {
	DAO dao.DAO
	Cs  cloudstorage.Interface
}

func (m *MCPFileHandler) DeleteFile(ctx context.Context, request *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	file, err := m.DAO.McpFilesystemClient().GetFileById(request.FileId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.DeleteFileResponse{
				Msg: fmt.Sprintf("no file Id:%v found, no need to delete", request.FileId),
			}, nil
		}
		logrus.Errorf("failed to get file by id: %v", err)
		return nil, errors.New("failed to get file by id")
	}

	if err = m.DAO.McpFilesystemClient().DeleteFile(file.ID); err != nil {
		logrus.Errorf("failed to delete file: %v", err)
		return nil, errors.New("failed to delete file")
	}

	if err = m.Cs.DeleteObject(ctx, file.ObjectKey); err != nil {
		logrus.Errorf("failed to delete oss object: %v", err)
		return nil, errors.New("failed to delete oss object")
	}

	return &pb.DeleteFileResponse{
		Msg: "success",
	}, nil
}

func (m *MCPFileHandler) UploadFile(ctx context.Context, request *pb.FileUploadRequest) (*pb.FileUploadResponse, error) {
	raw := ctx.Value(http.RequestContextKey)

	req, ok := raw.(*http1.Request)
	if !ok {
		return nil, errors.New("invalid request")
	}

	// 32MB
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	defer file.Close()

	var keep string = "Y"
	if req.FormValue("keep") == "" {
		keep = "N"
	}

	ext := filepath.Ext(header.Filename)
	filename := uuid.New()
	path := fmt.Sprintf("%s%s", filename, ext)

	result, err := m.Cs.PutObjectWithReader(ctx, file, path)
	if err != nil {
		logrus.Infof("failed to put object: %v", err)
		return nil, fmt.Errorf("failed to put object")
	}

	var endpoint, relationId string
	relation, err := m.DAO.McpFilesystemClient().GetRelationByBucketAndRegion(result.Bucket, result.Region)
	if err != nil || relation == nil {
		endpoint = defaultEndPoint(m.Cs.WhoIAm(), result.Bucket, result.ObjectName, result.Region)
		logrus.Warningf("failed to get relation by bucket: %v, use default endpoint: %s", err, endpoint)
	} else {
		relationId = relation.ID
		endpoint = fmt.Sprintf("https://%s", filepath.Join(relation.Domain, result.ObjectName))
	}

	id := uuid.New()
	if err = m.DAO.McpFilesystemClient().InsertFile(mcp_filesystem.McpFile{
		ID:          id,
		StorageType: m.Cs.WhoIAm(),
		ObjectKey:   result.ObjectName,
		FileName:    header.Filename,
		FileSize:    header.Size,
		FileMd5:     result.ContentMD5,
		ETag:        result.ETag,
		VersionID:   result.VersionId,
		Keep:        keep,
		IsDeleted:   "N",
		RelationId:  relationId,
	}); err != nil {
		logrus.Errorf("failed to save file info: %v", err)
	}

	return &pb.FileUploadResponse{
		FileId:     id,
		Url:        endpoint,
		UploadedAt: timestamppb.Now(),
	}, nil
}

func defaultEndPoint(storageType, bucket, objectName, region string) string {
	switch storageType {
	case "oss":
		return fmt.Sprintf("http://%s.oss-%s.aliyuncs.com/%s", bucket, region, objectName)
	}
	return ""
}
