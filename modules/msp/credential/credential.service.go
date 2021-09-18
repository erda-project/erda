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

package credential

import (
	"bytes"
	"context"
	"encoding/csv"

	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type accessKeyService struct {
	p *provider
}

const MSP_SCOPE = "msp_env"

func (a *accessKeyService) QueryAccessKeys(ctx context.Context, request *pb.QueryAccessKeysRequest) (*pb.QueryAccessKeysResponse, error) {
	req := &akpb.QueryAccessKeysRequest{
		Status:      request.Status,
		SubjectType: request.SubjectType,
		Subject:     request.Subject,
		AccessKey:   request.AccessKey,
		PageNo:      request.PageNo,
		PageSize:    request.PageSize,
		Scope:       MSP_SCOPE,
		ScopeId:     request.ScopeId,
	}
	accessKeyList, err := a.p.AccessKeyService.QueryAccessKeys(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	akList := make([]*pb.QueryAccessKeys, 0)
	for _, v := range accessKeyList.Data {
		ak := &pb.QueryAccessKeys{
			Id:        v.Id,
			Token:     v.AccessKey,
			CreatedAt: v.CreatedAt,
		}
		akList = append(akList, ak)
	}
	result := &pb.QueryAccessKeysResponse{
		Data: &pb.QueryAccessKeysData{
			List: akList,
		},
	}
	result.Data.Total = accessKeyList.Total
	return result, nil
}

func (a *accessKeyService) DownloadAccessKeyFile(ctx context.Context, request *pb.DownloadAccessKeyFileRequest) (*pb.DownloadAccessKeyFileResponse, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	akRequest := &akpb.GetAccessKeyRequest{
		Id: request.Id,
	}
	accessKey, err := a.p.AccessKeyService.GetAccessKey(ctx, akRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	fileData := [][]string{
		{"secretKey", accessKey.Data.SecretKey},
		{"accessKey", accessKey.Data.AccessKey},
	}
	err = w.WriteAll(fileData)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	w.Flush()
	return &pb.DownloadAccessKeyFileResponse{
		Content: buf.Bytes(),
	}, nil
}

func (a *accessKeyService) CreateAccessKey(ctx context.Context, request *pb.CreateAccessKeyRequest) (*pb.CreateAccessKeyResponse, error) {
	req := &akpb.CreateAccessKeyRequest{
		SubjectType: request.SubjectType,
		Subject:     request.Subject,
		Description: request.Description,
		Scope:       MSP_SCOPE,
		ScopeId:     request.ScopeId,
	}
	accessKey, err := a.p.AccessKeyService.CreateAccessKey(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.CreateAccessKeyResponse{
		Data: accessKey.Data.AccessKey,
	}
	return result, nil
}

func (a *accessKeyService) DeleteAccessKey(ctx context.Context, request *pb.DeleteAccessKeyRequest) (*pb.DeleteAccessKeyResponse, error) {
	akRequest := &akpb.DeleteAccessKeyRequest{
		Id: request.Id,
	}
	_, err := a.p.AccessKeyService.DeleteAccessKey(ctx, akRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *accessKeyService) GetAccessKey(ctx context.Context, request *pb.GetAccessKeyRequest) (*pb.GetAccessKeyResponse, error) {
	akRequest := &akpb.GetAccessKeyRequest{
		Id: request.Id,
	}
	accessKey, err := a.p.AccessKeyService.GetAccessKey(ctx, akRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetAccessKeyResponse{
		Data: accessKey.Data,
	}
	return result, nil
}
