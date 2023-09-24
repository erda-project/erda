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

package client_model_relation

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client"
)

type DBClient struct {
	DB           *gorm.DB
	ClientClient *client.DBClient
}

func (dbClient *DBClient) Allocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if err := TxCheckClientID(tx, req.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}
	// check model ids
	if err := TxCheckModelIDs(tx, req.ModelIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	// do allocate
	for _, modelId := range req.ModelIds {
		if err := (&models.AIProxyClientModelRelation{
			ClientID: req.ClientId,
			ModelID:  modelId,
		}).Creator(tx).Create(); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to allocate model %s to client %s: %v", modelId, req.ClientId, err)
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
	// check model ids
	if err := TxCheckModelIDs(tx, req.ModelIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	// do unallocate
	var relation models.AIProxyClientModelRelation
	if _, err := (&relation).Deleter(tx).Where(relation.FieldModelID().In(req.GetModelIds())).Delete(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to unallocate models %v from client %s: %v", req.ModelIds, req.ClientId, err)
	}
	tx.Commit()
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) ListClientModels(ctx context.Context, req *pb.ListClientModelsRequest) (*pb.ListAllocatedModelsResponse, error) {
	var relations models.AIProxyClientModelRelationList
	var wheres = []models.Where{
		relations.FieldClientID().Equal(req.GetClientId()),
	}
	if len(req.GetModelTypes()) > 0 {
		var models_ models.AIProxyModelList
		total, err := (&models_).Pager(dbClient.DB).Where(models_.FieldType().In(req.GetModelTypes())).Paging(-1, -1)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list models")
		}
		if total == 0 {
			return nil, errors.Wrap(gorm.ErrRecordNotFound, "failed to list models")
		}
		wheres = append(wheres, relations.FieldModelID().In(models_.FieldBaseModelList().FieldIDList()))
	}

	total, err := (&relations).Pager(dbClient.DB).Where(wheres...).Paging(-1, -1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list relations")
	}
	if total == 0 {
		return nil, errors.Wrap(gorm.ErrRecordNotFound, "failed to list relations")
	}

	return &pb.ListAllocatedModelsResponse{
		ClientId: req.ClientId,
		ModelIds: relations.FieldModelIDList(),
	}, nil
}

func TxCheckClientID(tx *gorm.DB, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client id is empty")
	}
	ok, err := (&models.AIProxyClient{}).Retriever(tx).Where(models.FieldID.Equal(clientID)).Get()
	if err != nil {
		return errors.Wrapf(err, "failed to check client id: %s", clientID)
	}
	if !ok {
		return errors.Wrapf(gorm.ErrRecordNotFound, "failed to check client id: %s", clientID)
	}

	return nil
}

func TxCheckModelIDs(tx *gorm.DB, modelIDs []string) error {
	if len(modelIDs) == 0 {
		return fmt.Errorf("modelIds is empty")
	}
	var models_ models.AIProxyModelList
	_, err := (&models_).Pager(tx).Where(models.FieldID.In(modelIDs)).Paging(-1, -1)
	if err != nil {
		return errors.Wrap(err, "failed to check model ids")
	}
	if existModelIds := models_.FieldBaseModelList().FieldIDStrList(); len(existModelIds) != len(modelIDs) {
		return fmt.Errorf("model ids %v not found", findMissingModelIds(modelIDs, existModelIds))
	}
	return nil
}

func findMissingModelIds(expected, actually []string) []string {
	actuallyMap := make(map[string]bool, len(actually))
	for _, id := range actually {
		actuallyMap[id] = true
	}
	var missings []string
	for _, expected := range expected {
		if _, ok := actuallyMap[expected]; !ok {
			missings = append(missings, expected)
		}
	}
	return missings
}
