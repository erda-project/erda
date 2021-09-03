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
	"fmt"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/transform_type"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type PipelineDefinitionProcess struct {
	ID                    uint64                              `json:"id"`
	PipelineSource        apistructs.PipelineSource           `json:"pipelineSource"`
	PipelineYmlName       string                              `json:"pipelineYmlName"`
	PipelineYml           string                              `json:"pipelineYml"`
	SnippetConfig         *apistructs.SnippetConfigOrder      `json:"snippetConfig"`
	VersionLock           uint64                              `json:"versionLock"`
	PipelineCreateRequest *apistructs.PipelineCreateRequestV2 `json:"pipelineCreateRequest"`

	IsDelete    bool       `json:"isDelete"`
	TimeCreated *time.Time `json:"timeCreated,omitempty"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty"`
}

type ProcessHandler func(definition PipelineDefinitionProcess, yml pipelineyml.PipelineYml) error

var definitionHandlers []ProcessHandler

func RegisterDefinitionHandler(definitionHandler ProcessHandler) {
	definitionHandlers = append(definitionHandlers, definitionHandler)
}

func definitionAdaptor(pipelineDefinition *PipelineDefinitionProcess) []error {
	pipelineYml, err := pipelineyml.New(
		[]byte(pipelineDefinition.PipelineYml),
	)
	if err != nil {
		return []error{err}
	}

	var wait sync.WaitGroup
	var handlerError []error
	for _, handler := range definitionHandlers {
		wait.Add(1)
		handler := handler
		go func() {
			defer wait.Done()
			handlerError = append(handlerError, handler(*pipelineDefinition, *pipelineYml))
		}()
	}
	wait.Wait()

	return handlerError
}

func (s *definitionService) ProcessPipelineDefinition(ctx context.Context, bo *transform_type.PipelineDefinitionProcess) (*transform_type.PipelineDefinitionProcessResult, error) {
	// validate
	if err := bo.Validate(); err != nil {
		return nil, apierrors.ErrrProcessPipelineDefinition.InvalidParameter(err)
	}

	dbPipelineDefinition, err := s.processPipelineDefinition(bo)
	if err != nil {
		return nil, err
	}

	errs := definitionAdaptor(&PipelineDefinitionProcess{
		ID:                    dbPipelineDefinition.ID,
		PipelineYml:           dbPipelineDefinition.PipelineYml,
		PipelineYmlName:       dbPipelineDefinition.PipelineYmlName,
		PipelineSource:        dbPipelineDefinition.PipelineSource,
		TimeCreated:           dbPipelineDefinition.TimeCreated,
		SnippetConfig:         dbPipelineDefinition.Extra.SnippetConfig,
		PipelineCreateRequest: dbPipelineDefinition.Extra.CreateRequest,
		TimeUpdated:           dbPipelineDefinition.TimeUpdated,
		IsDelete:              bo.IsDelete,
	})
	if len(errs) > 0 {
		return nil, apierrors.ErrrProcessPipelineDefinition.InvalidParameter(errs[0])
	}

	var result = &transform_type.PipelineDefinitionProcessResult{
		ID:              dbPipelineDefinition.ID,
		PipelineYml:     dbPipelineDefinition.PipelineYml,
		PipelineYmlName: dbPipelineDefinition.PipelineYmlName,
		PipelineSource:  dbPipelineDefinition.PipelineSource,
		VersionLock:     dbPipelineDefinition.VersionLock,
	}

	return result, nil
}

func (s *definitionService) processPipelineDefinition(req *transform_type.PipelineDefinitionProcess) (*db.PipelineDefinition, error) {
	dbPipelineDefinition, err := s.dbClient.GetPipelineDefinitionByNameAndSource(req.PipelineSource, req.PipelineYmlName)
	if err != nil {
		return nil, err
	}

	if dbPipelineDefinition == nil {
		dbPipelineDefinition = &db.PipelineDefinition{}
	}

	dbPipelineDefinition.PipelineYmlName = req.PipelineYmlName
	dbPipelineDefinition.PipelineSource = req.PipelineSource
	dbPipelineDefinition.PipelineYml = req.PipelineYml
	dbPipelineDefinition.Extra.SnippetConfig = req.SnippetConfig.Order()
	dbPipelineDefinition.Extra.CreateRequest = req.PipelineCreateRequest

	if req.IsDelete {
		err := s.dbClient.DeletePipelineDefinition(dbPipelineDefinition.ID)
		if err != nil {
			return nil, err
		}
		return dbPipelineDefinition, nil
	}

	if dbPipelineDefinition.ID > 0 {
		// check versionLock
		if dbPipelineDefinition.VersionLock != req.VersionLock {
			return nil, fmt.Errorf("db versionLock not match request versionLock")
		}
		dbPipelineDefinition.VersionLock = req.VersionLock
		err = s.dbClient.UpdatePipelineDefinition(dbPipelineDefinition.ID, dbPipelineDefinition)
	} else {
		// init versionLock
		dbPipelineDefinition.VersionLock = 1
		err = s.dbClient.CreatePipelineDefinition(dbPipelineDefinition)
	}
	if err != nil {
		return nil, err
	}

	return dbPipelineDefinition, nil
}
