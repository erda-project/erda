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

package session

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.SessionCreateRequest) (*pb.Session, error) {
	c := &Session{
		ClientID:    req.ClientId,
		PromptID:    req.PromptId,
		ModelID:     req.ModelId,
		Scene:       req.Scene,
		UserID:      req.UserId,
		Name:        req.Name,
		Topic:       req.Topic,
		NumOfCtxMsg: int64(req.NumOfCtxMsg),
		IsArchived:  false,
		Temperature: req.Temperature,
		Metadata:    metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.SessionGetRequest) (*pb.Session, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.SessionDeleteRequest) (*commonpb.VoidResponse, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.SessionUpdateRequest) (*pb.Session, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).Updates(Session{
		ClientID:    req.ClientId,
		PromptID:    req.PromptId,
		ModelID:     req.ModelId,
		Scene:       req.Scene,
		UserID:      req.UserId,
		Name:        req.Name,
		Topic:       req.Topic,
		NumOfCtxMsg: int64(req.NumOfCtxMsg),
		Temperature: req.Temperature,
		Metadata:    metadata.FromProtobuf(req.Metadata),
	}).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.SessionGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Archive(ctx context.Context, req *pb.SessionArchiveRequest) (*pb.Session, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id), IsArchived: true}
	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.SessionGetRequest{Id: req.Id})
}

func (dbClient *DBClient) UnArchive(ctx context.Context, req *pb.SessionUnArchiveRequest) (*pb.Session, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).Updates(map[string]any{"is_archived": false}).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.SessionGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Reset(ctx context.Context, req *pb.SessionResetRequest) (*pb.Session, error) {
	c := &Session{BaseModel: common.BaseModelWithID(req.Id), ResetAt: fields.DeletedAt{Time: time.Now(), Valid: true}}
	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.SessionGetRequest{Id: req.Id})
}

func (dbClient *DBClient) GetChatLogs(ctx context.Context, req *pb.SessionChatLogGetRequest) (*pb.SessionChatLogGetResponse, error) {
	var (
		count    int64
		chatLogs audit.Audits
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	c := &audit.Audit{SessionID: req.SessionId}
	if err := dbClient.DB.Model(c).Where(c).
		Count(&count).
		Limit(int(req.PageSize)).Offset(int(offset)).Find(&chatLogs).Error; err != nil {
		return nil, err
	}
	return &pb.SessionChatLogGetResponse{
		Total: uint64(count),
		List:  chatLogs.ToChatLogsProtobuf(),
	}, nil
}
