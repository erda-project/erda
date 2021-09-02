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

package credential

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type accessKeyService struct {
	p *provider
}

func (a *accessKeyService) QueryAccessKeys(ctx context.Context, request *pb.QueryAccessKeysRequest) (*pb.QueryAccessKeysResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &akpb.QueryAccessKeysRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	accessKeyList, err := a.p.AccessKeyService.QueryAccessKeys(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err = json.Marshal(accessKeyList.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryAccessKeysResponse{
		Data: &pb.QueryAccessKeysData{
			List: make([]*akpb.AccessKeysItem, 0),
		},
	}
	err = json.Unmarshal(data, &result.Data.List)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result.Data.Total = accessKeyList.Total
	return result, nil
}

func (a *accessKeyService) DownloadAccessKeyFile(ctx context.Context, request *pb.DownloadAccessKeyFileRequest) (*pb.DownloadAccessKeyFileResponse, error) {
	//先生成csv文件
	f, err := os.Create("accessKey.csv")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	//防止中文乱码
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	akRequest := &akpb.GetAccessKeyRequest{
		Id: request.Id,
	}
	accessKey, err := a.p.AccessKeyService.GetAccessKey(ctx, akRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	akMap := make(map[string]interface{})
	data, err := json.Marshal(accessKey.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, &akMap)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	fileData := make([][]string, 0)
	for k, v := range akMap {
		fileData = append(fileData, []string{k, fmt.Sprint(v)})
	}
	w.WriteAll(fileData)
	w.Flush()
	//返回
	return &pb.DownloadAccessKeyFileResponse{}, nil

}

func (a *accessKeyService) CreateAccessKey(ctx context.Context, request *pb.CreateAccessKeyRequest) (*pb.CreateAccessKeyResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &akpb.CreateAccessKeyRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	accessKey, err := a.p.AccessKeyService.CreateAccessKey(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.CreateAccessKeyResponse{
		Data: accessKey.Data,
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
