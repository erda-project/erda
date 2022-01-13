package dbclient

import (
	"fmt"

	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
)

func (client *Client) GetPipelineDefinition(id string, ops ...SessionOption) (*db.PipelineDefinition, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineDefinition db.PipelineDefinition
	var has bool
	var err error
	if has, _, err = session.Where("id = ? and soft_deleted_at = 0", id).GetFirst(&pipelineDefinition).GetResult(); err != nil {
		return nil, err
	}

	if !has {
		return nil, fmt.Errorf("the record not fount")
	}

	return &pipelineDefinition, nil
}

func (client *Client) UpdatePipelineDefinition(id string, pipelineDefinition *db.PipelineDefinition, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineDefinition)
	return err
}
