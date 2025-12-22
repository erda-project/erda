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

package client_token

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

const (
	TokenPrefix          = "t_"
	defaultExpireInHours = 24 * 7 // 7 days
)

func genToken() string {
	return TokenPrefix + uuid.UUID()
}

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ClientTokenCreateRequest) (*pb.ClientToken, error) {
	// set default expire option
	if req.ExpireInHours == 0 {
		req.ExpireInHours = defaultExpireInHours
	}
	// query first
	if req.CreateOrGet {
		pagingResp, err := dbClient.Paging(ctx, &pb.ClientTokenPagingRequest{
			ClientId: req.ClientId,
			UserId:   req.UserId,
			PageNum:  1,
			PageSize: 1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to paging token, clientId: %s, userId: %s, err: %v", req.ClientId, req.UserId, err)
		}
		if pagingResp.Total == 1 { // already exist, return directly
			// do update to auto-renewal token
			return dbClient.Update(ctx, &pb.ClientTokenUpdateRequest{
				ClientId:      req.ClientId,
				Token:         pagingResp.List[0].Token,
				Metadata:      req.Metadata,
				ExpireInHours: req.ExpireInHours,
			})
		}
	}
	// do create
	c := &ClientToken{
		ClientID:  req.ClientId,
		UserID:    req.UserId,
		Token:     genToken(),
		ExpiredAt: time.Now().Add(time.Hour * time.Duration(req.ExpireInHours)),
		Metadata:  metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ClientTokenGetRequest) (*pb.ClientToken, error) {
	c := &ClientToken{ClientID: req.ClientId, Token: req.Token}
	if err := dbClient.DB.Model(c).Where(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ClientTokenDeleteRequest) (*commonpb.VoidResponse, error) {
	tokenResp, err := dbClient.Get(ctx, &pb.ClientTokenGetRequest{ClientId: req.ClientId, Token: req.Token})
	if err != nil {
		return nil, fmt.Errorf("failed to get token, clientId: %s, token: %s, err:, %v", req.ClientId, req.Token, err)
	}
	c := &ClientToken{BaseModel: common.BaseModelWithID(tokenResp.Id), Token: req.Token}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ClientTokenUpdateRequest) (*pb.ClientToken, error) {
	tokenResp, err := dbClient.Get(ctx, &pb.ClientTokenGetRequest{ClientId: req.ClientId, Token: req.Token})
	if err != nil {
		return nil, fmt.Errorf("failed to get token, clientId: %s, token: %s, err:, %v", req.ClientId, req.Token, err)
	}
	c := &ClientToken{
		BaseModel: common.BaseModelWithID(tokenResp.Id),
	}
	if req.ExpireInHours > 0 {
		c.ExpiredAt = time.Now().Add(time.Hour * time.Duration(req.ExpireInHours))
	}
	if req.Metadata != nil {
		c.Metadata = metadata.FromProtobuf(req.Metadata)
	}
	sql := dbClient.DB.Model(c).Updates(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return dbClient.Get(ctx, &pb.ClientTokenGetRequest{ClientId: req.ClientId, Token: req.Token})
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ClientTokenPagingRequest) (*pb.ClientTokenPagingResponse, error) {
	c := &ClientToken{ClientID: req.ClientId, UserID: req.UserId, Token: req.Token}
	sql := dbClient.DB.Model(c)
	var (
		total int64
		list  ClientTokens
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	err := sql.WithContext(ctx).Where(c).Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &pb.ClientTokenPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
