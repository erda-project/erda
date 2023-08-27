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

package prompt

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_model_relation"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.PromptCreateRequest) (*pb.Prompt, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if req.ClientId != "" {
		if err := client_model_relation.TxCheckClientID(tx, req.ClientId); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// create
	c := &Prompt{
		Name:     req.Name,
		Desc:     req.Desc,
		ClientID: req.ClientId,
		Messages: message.FromProtobuf(req.Messages),
		Metadata: metadata.FromProtobuf(req.Metadata),
	}
	if err := tx.Model(c).Create(c).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	// commit
	tx.Commit()
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.PromptGetRequest) (*pb.Prompt, error) {
	c := &Prompt{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.PromptDeleteRequest) (*commonpb.VoidResponse, error) {
	c := &Prompt{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.PromptUpdateRequest) (*pb.Prompt, error) {
	c := &Prompt{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).Updates(Prompt{
		Name:     req.Name,
		Desc:     req.Desc,
		ClientID: req.ClientId,
		Messages: message.FromProtobuf(req.Messages),
		Metadata: metadata.FromProtobuf(req.Metadata),
	}).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.PromptGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.PromptPagingRequest) (*pb.PromptPagingResponse, error) {
	c := &Prompt{ClientID: req.ClientId}
	sql := dbClient.DB.Model(c).Where(c)
	if req.Name != "" {
		sql.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var (
		total int64
		list  Prompts
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	if err := sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error; err != nil {
		return nil, err
	}
	return &pb.PromptPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
