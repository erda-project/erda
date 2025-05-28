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
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/sqlutil"
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
	c := &Client{
		Name:        req.Name,
		Desc:        req.Desc,
		AccessKeyID: req.AccessKeyId,
		SecretKeyID: req.SecretKeyId,
		Metadata:    metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ClientGetRequest) (*pb.Client, error) {
	c := &Client{BaseModel: common.BaseModelWithID(req.ClientId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ClientDeleteRequest) (*commonpb.VoidResponse, error) {
	c := &Client{BaseModel: common.BaseModelWithID(req.ClientId)}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ClientUpdateRequest) (*pb.Client, error) {
	c := &Client{
		BaseModel:   common.BaseModelWithID(req.ClientId),
		Name:        req.Name,
		Desc:        req.Desc,
		AccessKeyID: req.AccessKeyId,
		SecretKeyID: req.SecretKeyId,
		Metadata:    metadata.FromProtobuf(req.Metadata),
	}
	sql := dbClient.DB.Model(c).Updates(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return dbClient.Get(ctx, &pb.ClientGetRequest{ClientId: req.ClientId})
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ClientPagingRequest) (*pb.ClientPagingResponse, error) {
	c := &Client{}
	sql := dbClient.DB.Model(c)
	if req.Name != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id IN (?)", req.Ids)
	}
	if len(req.AccessKeyIds) > 0 {
		sql = sql.Where("access_key_id IN (?)", req.AccessKeyIds)
	}
	// order by
	sql, err := sqlutil.HandleOrderBy(sql, req.OrderBys)
	if err != nil {
		return nil, err
	}
	var (
		total int64
		list  Clients
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	err = sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &pb.ClientPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
