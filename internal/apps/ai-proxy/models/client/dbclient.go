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

package client

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ClientCreateRequest) (*pb.Client, error) {
	if req.AccessKeyId == "" {
		req.AccessKeyId = uuid.UUID()
	}
	if req.SecretKeyId == "" {
		req.SecretKeyId = uuid.UUID()
	}
	c := &models.AIProxyClient{
		Name:        req.Name,
		Desc:        req.Desc,
		AccessKeyID: req.AccessKeyId,
		SecretKeyID: req.SecretKeyId,
		Metadata:    models.MetadataFromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ClientGetRequest) (*pb.Client, error) {
	var c models.AIProxyClient
	ok, err := (&c).Retriever(dbClient.DB).Where(c.FieldID().Equal(req.GetClientId())).Get()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ClientDeleteRequest) (*commonpb.VoidResponse, error) {
	affects, err := (&models.AIProxyClient{}).Deleter(dbClient.DB).Where(models.FieldID.Equal(req.GetClientId())).Delete()
	if err != nil {
		return nil, err
	}
	if affects < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return new(commonpb.VoidResponse), nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ClientUpdateRequest) (*pb.Client, error) {
	var client models.AIProxyClient
	affects, err := (&client).Updater(dbClient.DB).
		Where(models.FieldID.Equal(req.GetClientId())).
		Updates(
			client.FieldName().Set(req.GetName()),
			client.FieldDesc().Set(req.GetDesc()),
			client.FieldAccessKeyID().Set(req.GetAccessKeyId()),
			client.FieldSecretKeyID().Set(req.GetSecretKeyId()),
			client.FieldMetadata().Set(models.MetadataFromProtobuf(req.GetMetadata())),
		)
	if err != nil {
		return nil, err
	}
	if affects < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return client.ToProtobuf(), nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ClientPagingRequest) (*pb.ClientPagingResponse, error) {
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	var list models.AIProxyClientList
	var wheres []models.Where
	if req.GetName() != "" {
		wheres = append(wheres, list.FieldName().Like("%"+req.Name+"%"))
	}
	if len(req.GetIds()) > 0 {
		wheres = append(wheres, models.FieldID.In(req.GetIds()))
	}
	if len(req.GetAccessKeyIds()) > 0 {
		wheres = append(wheres, list.FieldAccessKeyID().In(req.GetAccessKeyIds()))
	}

	total, err := (&list).Pager(dbClient.DB).Where(wheres...).Paging(int(req.GetPageSize()), int(req.GetPageNum()))
	if err != nil {
		return nil, err
	}
	return &pb.ClientPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
