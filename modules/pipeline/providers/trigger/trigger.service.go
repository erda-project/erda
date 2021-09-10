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

package trigger

import (
	context "context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/trigger/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition"
	definitionDb "github.com/erda-project/erda/modules/pipeline/providers/definition/db"
	triggerDb "github.com/erda-project/erda/modules/pipeline/providers/trigger/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type TriggerService struct {
	p *provider

	db                 mysqlxorm.Interface
	triggerDbClient    *triggerDb.Client
	definitionDbClient *definitionDb.Client
	pipelineSvc        *pipelinesvc.PipelineSvc
}

func (s *TriggerService) SetPipelineSvc(pipelineSvc *pipelinesvc.PipelineSvc) {
	s.pipelineSvc = pipelineSvc
}

func (s *TriggerService) RunPipelineByTriggerRequest(ctx context.Context, req *pb.PipelineTriggerRequest) (*pb.PipelineTriggerResponse, error) {
	err := s.checkPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCheckPermission.AccessDenied()
	}

	pipelineTriggers, err := s.triggerDbClient.ListPipelineTriggers(req)
	if err != nil {
		return nil, err
	}
	pipelineTriggersMap := GetUniqueTrigger(pipelineTriggers)

	// Get trigger label
	triggerLabel := make(map[string]string)
	for key, val := range req.Label {
		triggerLabel[fmt.Sprintf("%s.%s.%s", expression.TriggerLabel, req.EventName, key)] = val
	}

	var pipelineIDs []uint64
	for _, trigger := range pipelineTriggersMap {

		pipelineDefinition, err := s.definitionDbClient.GetPipelineDefinitionByNameAndSource(trigger.PipelineSource, trigger.PipelineYmlName)
		if err != nil {
			return nil, err
		}

		if pipelineDefinition.PipelineYmlName == "" || pipelineDefinition.PipelineYml == "" || pipelineDefinition.PipelineSource == "" || pipelineDefinition.Extra.CreateRequest == nil {
			continue
		}

		label := make(map[string]string)
		if pipelineDefinition.Extra.CreateRequest != nil && pipelineDefinition.Extra.CreateRequest.Labels != nil {
			label = pipelineDefinition.Extra.CreateRequest.Labels
		}
		for key, val := range triggerLabel {
			label[key] = val
		}

		pipeline, err := s.pipelineSvc.CreateV2(&apistructs.PipelineCreateRequestV2{
			PipelineYml:            pipelineDefinition.PipelineYml,
			ClusterName:            pipelineDefinition.Extra.CreateRequest.ClusterName,
			PipelineYmlName:        pipelineDefinition.PipelineYmlName,
			PipelineSource:         pipelineDefinition.PipelineSource,
			Labels:                 label,
			NormalLabels:           pipelineDefinition.Extra.CreateRequest.NormalLabels,
			ConfigManageNamespaces: pipelineDefinition.Extra.CreateRequest.ConfigManageNamespaces,
			AutoRunAtOnce:          true,
			AutoStartCron:          false,
			ForceRun:               false,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: apis.GetUserID(ctx),
			},
		})
		if err != nil {
			return nil, err
		}
		pipelineIDs = append(pipelineIDs, pipeline.ID)
	}

	return &pb.PipelineTriggerResponse{PipelineIDs: pipelineIDs}, nil
}

// TODO: add original pipeline yml
func (s *TriggerService) RegisterTriggerHandler(definition definition.PipelineDefinitionProcess, yml pipelineyml.PipelineYml) error {

	newPipelineTriggerMap := GetPipelineTriggerMap(yml)

	oldPipelineTriggers, err := s.triggerDbClient.GetPipelineTriggerByID(definition.ID)
	if err != nil {
		return err
	}

	txSession := s.triggerDbClient.NewSession()
	for _, oldPipelineTrigger := range oldPipelineTriggers {
		err := s.triggerDbClient.DeletePipelineTrigger(oldPipelineTrigger.ID, mysqlxorm.WithSession(txSession))
		if err != nil {
			return err
		}
	}

	if definition.IsDelete {
		return nil
	}

	for eventName, newPipelineTrigger := range newPipelineTriggerMap {
		err := s.triggerDbClient.CreatePipelineTrigger(&triggerDb.PipelineTrigger{
			ID:                   uuid.SnowFlakeIDUint64(),
			Event:                eventName,
			PipelineSource:       definition.PipelineSource,
			PipelineYmlName:      definition.PipelineYmlName,
			PipelineDefinitionID: definition.ID,
			Filter:               newPipelineTrigger,
		}, mysqlxorm.WithSession(txSession))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TriggerService) checkPermission(ctx context.Context) error {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return errors.Errorf("failed to get permission (User-ID is empty)")
	}
	orgID := apis.GetOrgID(ctx)
	if orgID == "" {
		return errors.Errorf("failed to get permission (Org-ID is empty)")
	}
	return nil
}

func GetPipelineTriggerMap(yml pipelineyml.PipelineYml) map[string]map[string]string {
	newPipelineTriggerMap := make(map[string]map[string]string)
	if yml.Spec() != nil && yml.Spec().Triggers != nil {
		if yml.Spec().Triggers != nil {
			for _, trigger := range yml.Spec().Triggers {
				if trigger.Filter != nil {
					newPipelineTriggerMap[trigger.On] = trigger.Filter
				}
			}
		}
	}
	return newPipelineTriggerMap
}

func GetUniqueTrigger(pipelineTriggers []triggerDb.PipelineTrigger) map[uint64]triggerDb.PipelineTrigger {
	pipelineTriggersMap := make(map[uint64]triggerDb.PipelineTrigger)
	for _, trigger := range pipelineTriggers {
		if _, ok := pipelineTriggersMap[trigger.PipelineDefinitionID]; !ok {
			pipelineTriggersMap[trigger.PipelineDefinitionID] = trigger
		}
	}
	return pipelineTriggersMap
}
