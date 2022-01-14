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

package definition

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/time/mysql_time"
)

type pipelineDefinition struct {
	dbClient *db.Client
}

func (p pipelineDefinition) Create(ctx context.Context, request *pb.PipelineDefinitionCreateRequest) (*pb.PipelineDefinitionCreateResponse, error) {
	if err := createPreCheck(request); err != nil {
		return nil, err
	}

	var pipelineDefinition db.PipelineDefinition
	pipelineDefinition.PipelineSourceId = request.PipelineSourceId
	pipelineDefinition.Category = request.Category
	pipelineDefinition.Creator = request.Creator
	pipelineDefinition.Name = request.Name
	pipelineDefinition.ID = uuid.New().String()
	pipelineDefinition.StartedAt = *mysql_time.GetMysqlDefaultTime()
	pipelineDefinition.EndedAt = *mysql_time.GetMysqlDefaultTime()
	err := p.dbClient.CreatePipelineDefinition(&pipelineDefinition)
	if err != nil {
		return nil, err
	}

	var pipelineDefinitionExtra db.PipelineDefinitionExtra
	pipelineDefinitionExtra.ID = uuid.New().String()
	var extra apistructs.PipelineDefinitionExtraValue
	err = json.Unmarshal([]byte(request.Extra.Extra), &extra)
	if err != nil {
		return nil, err
	}
	pipelineDefinitionExtra.PipelineDefinitionID = pipelineDefinition.ID
	pipelineDefinitionExtra.Extra = extra
	err = p.dbClient.CreatePipelineDefinitionExtra(&pipelineDefinitionExtra)
	if err != nil {
		return nil, err
	}

	pbPipelineDefinition := PipelineDefinitionToPb(&pipelineDefinition)
	pbPipelineDefinitionExtra := PipelineDefinitionExtraToPb(&pipelineDefinitionExtra)
	pbPipelineDefinition.Extra = pbPipelineDefinitionExtra
	return &pb.PipelineDefinitionCreateResponse{
		PipelineDefinition: pbPipelineDefinition,
	}, nil
}

func createPreCheck(request *pb.PipelineDefinitionCreateRequest) error {
	if request.Name == "" || len(request.Name) > 36 {
		return apierrors.ErrCreatePipelineDefinition.InvalidParameter(errors.Errorf("name: %s", request.Name))
	}
	if request.Creator == "" {
		return apierrors.ErrCreatePipelineDefinition.InvalidParameter(errors.Errorf("creator: %s", request.Creator))
	}
	if request.Category == "" {
		return apierrors.ErrCreatePipelineDefinition.InvalidParameter(errors.Errorf("category: %s", request.Category))
	}
	if request.PipelineSourceId == "" {
		return apierrors.ErrCreatePipelineDefinition.InvalidParameter(errors.Errorf("pipelineSourceId: %s", request.PipelineSourceId))
	}
	if request.Extra == nil || request.Extra.Extra == "" {
		return apierrors.ErrCreatePipelineDefinition.InvalidParameter(errors.Errorf("extra: %s", request.Extra))
	}
	return nil
}

func (p pipelineDefinition) Update(ctx context.Context, request *pb.PipelineDefinitionUpdateRequest) (*pb.PipelineDefinitionUpdateResponse, error) {
	if request.PipelineDefinitionID == "" {
		return nil, apierrors.ErrUpdatePipelineDefinition.InvalidParameter(errors.Errorf("pipelineDefinitionID: %s", request.PipelineDefinitionID))
	}
	pipelineDefinition, err := p.dbClient.GetPipelineDefinition(request.PipelineDefinitionID)
	if err != nil {
		return nil, err
	}
	if request.Category != "" {
		pipelineDefinition.Category = request.Category
	}
	if request.Name != "" {
		pipelineDefinition.Name = request.Name
	}
	if request.CostTime > 0 {
		pipelineDefinition.CostTime = request.CostTime
	}
	if request.PipelineSourceId != "" {
		pipelineDefinition.PipelineSourceId = request.PipelineSourceId
	}
	if request.StartedAt != nil {
		var startAt = request.StartedAt.AsTime()
		pipelineDefinition.StartedAt = startAt
	}
	if request.EndedAt != nil {
		var endAt = request.EndedAt.AsTime()
		pipelineDefinition.EndedAt = endAt
	}
	if request.Executor != "" {
		pipelineDefinition.Executor = request.Executor
	}
	if request.Status != "" {
		pipelineDefinition.Status = request.Status
	}
	if request.PipelineId > 0 {
		pipelineDefinition.PipelineID = uint64(request.PipelineId)
	}
	err = p.dbClient.UpdatePipelineDefinition(request.PipelineDefinitionID, pipelineDefinition)
	if err != nil {
		return nil, err
	}

	pbPipelineDefinition := PipelineDefinitionToPb(pipelineDefinition)
	return &pb.PipelineDefinitionUpdateResponse{
		PipelineDefinition: pbPipelineDefinition,
	}, nil
}

func (p pipelineDefinition) Delete(ctx context.Context, request *pb.PipelineDefinitionDeleteRequest) (*pb.PipelineDefinitionDeleteResponse, error) {
	err := p.dbClient.DeletePipelineDefinition(request.PipelineDefinitionID)
	if err != nil {
		return nil, err
	}

	return &pb.PipelineDefinitionDeleteResponse{}, nil
}

func (p pipelineDefinition) Get(ctx context.Context, request *pb.PipelineDefinitionGetRequest) (*pb.PipelineDefinitionGetResponse, error) {
	pipelineDefinition, err := p.dbClient.GetPipelineDefinition(request.PipelineDefinitionID)
	if err != nil {
		return nil, err
	}

	pipelineDefinitionExtra, err := p.dbClient.GetPipelineDefinitionExtraByDefinitionID(pipelineDefinition.ID)
	if err != nil {
		return nil, err
	}

	pbPipelineDefinition := PipelineDefinitionToPb(pipelineDefinition)
	pbPipelineDefinitionExtra := PipelineDefinitionExtraToPb(pipelineDefinitionExtra)
	pbPipelineDefinition.Extra = pbPipelineDefinitionExtra
	return &pb.PipelineDefinitionGetResponse{
		PipelineDefinition: pbPipelineDefinition,
	}, nil
}

func (p pipelineDefinition) List(ctx context.Context, request *pb.PipelineDefinitionListRequest) (*pb.PipelineDefinitionListResponse, error) {
	definitions, total, err := p.dbClient.ListPipelineDefinition(request)
	if err != nil {
		return nil, err
	}

	data := make([]*pb.PipelineDefinition, 0, len(definitions))
	var definitionIDList []string
	for _, v := range definitions {
		definitionIDList = append(definitionIDList, v.ID)
		data = append(data, v.Convert())
	}

	var extrasMap = map[string]db.PipelineDefinitionExtra{}
	extras, err := p.dbClient.ListPipelineDefinitionExtraByDefinitionIDList(definitionIDList)
	if err != nil {
		return nil, err
	}
	for _, extra := range extras {
		extrasMap[extra.PipelineDefinitionID] = extra
	}

	for _, definition := range data {
		definition.Extra = &pb.PipelineDefinitionExtra{
			ID:    extrasMap[definition.ID].ID,
			Extra: jsonparse.JsonOneLine(extrasMap[definition.ID].Extra),
		}
	}

	return &pb.PipelineDefinitionListResponse{
		Total: total,
		Data:  data,
	}, nil
}

func PipelineDefinitionToPb(pipelineDefinition *db.PipelineDefinition) *pb.PipelineDefinition {
	de := &pb.PipelineDefinition{
		ID:               pipelineDefinition.ID,
		Name:             pipelineDefinition.Name,
		Creator:          pipelineDefinition.Creator,
		Executor:         pipelineDefinition.Executor,
		CostTime:         pipelineDefinition.CostTime,
		Category:         pipelineDefinition.Category,
		PipelineSourceId: pipelineDefinition.PipelineSourceId,
		TimeCreated:      timestamppb.New(pipelineDefinition.TimeCreated),
		TimeUpdated:      timestamppb.New(pipelineDefinition.TimeUpdated),
		StartedAt:        timestamppb.New(pipelineDefinition.StartedAt),
		EndedAt:          timestamppb.New(pipelineDefinition.EndedAt),
	}
	return de
}

func PipelineDefinitionExtraToPb(pipelineDefinitionExtra *db.PipelineDefinitionExtra) *pb.PipelineDefinitionExtra {
	de := &pb.PipelineDefinitionExtra{
		ID:    pipelineDefinitionExtra.ID,
		Extra: jsonparse.JsonOneLine(pipelineDefinitionExtra.Extra),
	}
	if pipelineDefinitionExtra.TimeCreated != nil {
		de.TimeCreated = timestamppb.New(*pipelineDefinitionExtra.TimeCreated)
	}
	if pipelineDefinitionExtra.TimeUpdated != nil {
		de.TimeUpdated = timestamppb.New(*pipelineDefinitionExtra.TimeUpdated)
	}
	return de
}
