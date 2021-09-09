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

package db

import (
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/trigger/pb"
)

type Client struct {
	mysqlxorm.Interface
}

func (client *Client) CreatePipelineTrigger(trigger *PipelineTrigger, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.InsertOne(trigger)
	return err
}

func (client *Client) UpdatePipelineTrigger(id uint64, trigger *PipelineTrigger, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(trigger)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) DeletePipelineTrigger(id uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if _, err := session.Where("id = ?", id).Delete(&PipelineTrigger{}); err != nil {
		return err
	}
	return nil
}

func (client *Client) ListPipelineTriggers(req *pb.PipelineTriggerRequest, ops ...mysqlxorm.SessionOption) ([]PipelineTrigger, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var triggers []PipelineTrigger
	if err := session.In("event", req.EventName).Find(&triggers); err != nil {
		return nil, err
	}
	filterTriggers, err := FilterByEvent(triggers, req.Label)
	if err != nil {
		return nil, err
	}
	return filterTriggers, nil
}

func (client *Client) GetPipelineTriggerByID(pipelineDefinitionID uint64, ops ...mysqlxorm.SessionOption) ([]PipelineTrigger, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var triggers []PipelineTrigger
	if err := session.In("pipeline_definition_id", pipelineDefinitionID).Find(&triggers); err != nil {
		return nil, err
	}
	return triggers, nil
}

func FilterByEvent(triggers []PipelineTrigger, Filter map[string]string) ([]PipelineTrigger, error) {
	var filterTriggers []PipelineTrigger
	for _, trigger := range triggers {
		isFilterTrigger := true
		for k, v := range trigger.Filter {
			filterVal, ok := Filter[k]
			if !ok {
				isFilterTrigger = false
				break
			}
			if filterVal != v {
				isFilterTrigger = false
				break
			}
		}
		if isFilterTrigger == true {
			filterTriggers = append(filterTriggers, trigger)
		}
	}
	return filterTriggers, nil
}
