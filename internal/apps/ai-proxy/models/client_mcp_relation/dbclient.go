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

package client_mcp_relation

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_mcp_relation/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

type DBClient struct {
	DB           *gorm.DB
	ClientClient *client.DBClient
}

func (dbClient *DBClient) ListClientMCPScope(ctx context.Context, request *pb.ListClientMCPScopeRequest) (*pb.ListAllocatedMCPScopeResponse, error) {
	tx := dbClient.DB.Begin()
	var relations []*ClientMcpRelation

	if err := TxCheckClientID(tx, request.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Model(&ClientMcpRelation{}).Where("client_id = ?", request.ClientId).Find(&relations).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var resp pb.ListAllocatedMCPScopeResponse
	resp.Scope = make(map[string]*pb.ScopeIdList)
	for _, relation := range relations {
		if _, ok := resp.Scope[relation.ScopeType]; !ok {
			resp.Scope[relation.ScopeType] = &pb.ScopeIdList{
				Ids: make([]string, 0),
			}
		}

		resp.Scope[relation.ScopeType].Ids = append(resp.Scope[relation.ScopeType].Ids, relation.ScopeID)
	}

	return &resp, nil
}

func (dbClient *DBClient) Allocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if err := TxCheckClientID(tx, req.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}
	// do allocate
	for _, id := range req.ScopeIds {
		c := &ClientMcpRelation{
			ClientID:  req.ClientId,
			ScopeType: req.ScopeType,
			ScopeID:   id,
		}
		if err := tx.Model(c).Create(c).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to allocate mcp scopeType %s, scopeId %s to client %s: %v", req.ScopeType, id, req.ClientId, err)
		}
	}
	tx.Commit()
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) UnAllocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if err := TxCheckClientID(tx, req.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}

	// do unallocate
	if err := tx.Model(&ClientMcpRelation{ClientID: req.ClientId}).
		Where("scope_id in (?)", req.ScopeIds).
		Where("scope_type = ?", req.ScopeType).
		Delete(nil).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to unallocate mcp scopeType %s, scopeIds %v from client %s: %v", req.ScopeType, req.ScopeIds, req.ClientId, err)
	}
	tx.Commit()
	return &commonpb.VoidResponse{}, nil
}

func TxCheckClientID(tx *gorm.DB, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client id is empty")
	}
	var clientCount int64
	c := &client.Client{BaseModel: common.BaseModelWithID(clientID)}
	if err := tx.Model(c).Where(c).Count(&clientCount).Error; err != nil {
		return fmt.Errorf("failed to check client id: %v", err)
	}
	if clientCount == 0 {
		return fmt.Errorf("client id %s not found", clientID)
	}
	return nil
}
