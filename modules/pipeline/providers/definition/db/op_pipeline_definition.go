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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

func (client *Client) CreatePipelineDefinition(pipelineDefinition *PipelineDefinition, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err = session.InsertOne(pipelineDefinition)
	return err
}

func (client *Client) UpdatePipelineDefinition(id uint64, pipelineDefinition *PipelineDefinition, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineDefinition)
	return err
}

func (client *Client) DeletePipelineDefinition(id uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).Delete(new(PipelineDefinition))
	return err
}

func (client *Client) GetPipelineDefinitionByNameAndSource(PipelineSource apistructs.PipelineSource, PipelineYmlName string, ops ...mysqlxorm.SessionOption) (*PipelineDefinition, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineDefinition PipelineDefinition
	var has bool
	var err error
	if has, _, err = session.Where("pipeline_source = ? and pipeline_yml_name = ?", PipelineSource, PipelineYmlName).GetFirst(&pipelineDefinition).GetResult(); err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &pipelineDefinition, nil
}

func (client *Client) GetPipelineDefinitionBySnippetConfigOrder(snippetConfig apistructs.SnippetConfigOrder, ops ...mysqlxorm.SessionOption) (*PipelineDefinition, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineDefinition PipelineDefinition
	if err := session.Where("snippet_config = ?", jsonparse.JsonOneLine(snippetConfig)).GetFirst(&pipelineDefinition).Error; err != nil {
		return nil, err
	}

	return &pipelineDefinition, nil
}
