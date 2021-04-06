// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dbclient

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) ListPipelineBasesByIDs(pipelineIDs []uint64, ops ...SessionOption) (map[uint64]spec.PipelineBase, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var bases []spec.PipelineBase
	if err := session.In("id", pipelineIDs).Find(&bases); err != nil {
		return nil, err
	}
	basesMap := make(map[uint64]spec.PipelineBase, len(bases))
	for _, base := range bases {
		basesMap[base.ID] = base
	}
	return basesMap, nil
}

func (client *Client) CreatePipelineBase(base *spec.PipelineBase, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.InsertOne(base)
	return err
}

func (client *Client) UpdatePipelineBase(id uint64, base *spec.PipelineBase, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(base)
	return err
}

func (client *Client) GetPipelineBase(id uint64, ops ...SessionOption) (spec.PipelineBase, bool, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var base spec.PipelineBase
	found, err := session.ID(id).Get(&base)
	if err != nil {
		return spec.PipelineBase{}, false, err
	}
	return base, found, nil
}

func (client *Client) GetPipelineStatus(id uint64, ops ...SessionOption) (apistructs.PipelineStatus, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var base spec.PipelineBase
	exist, err := session.ID(id).Cols("status").Get(&base)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", fmt.Errorf("pipeline base not found")
	}
	return base.Status, nil
}

func (client *Client) UpdatePipelineBaseStatus(id uint64, status apistructs.PipelineStatus, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).Cols("status").Update(&spec.PipelineBase{Status: status})
	return err
}
