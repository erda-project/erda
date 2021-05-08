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

	"gorm.io/gorm"

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

// use insert into on duplicate key update to batch update, id absolutely exit, so it always update
func (client *Client) BatchUpdatePipelineBaseParentID(bases []*spec.PipelineBase, tx *gorm.DB) (err error) {
	if len(bases) <= 0 {
		return nil
	}

	var gormClient = tx
	if gormClient == nil {
		gormClient = client.DB
	}

	updateSql := "UPDATE pipeline_bases set "

	var baseIndex = 0
	var batchInsertNum = 4000 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(bases) / batchInsertNum

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(bases) - batchInsertNum*sqlWillRunTimes
		}

		var sqlArgs []interface{}

		var setParentTaskIndex = baseIndex
		var setParentTaskSql = " parent_task_id = case id "
		for index := 0; index < forNum; index++ {
			base := bases[setParentTaskIndex]
			if base.ID <= 0 {
				continue
			}
			setParentTaskSql += " when ? then ? "

			sqlArgs = append(sqlArgs, base.ID)
			sqlArgs = append(sqlArgs, base.ParentTaskID)
			setParentTaskIndex++
		}

		var setParentPipelineIndex = baseIndex
		var setParentPipelineSql = " parent_pipeline_id = case id "
		for index := 0; index < forNum; index++ {
			base := bases[setParentPipelineIndex]
			if base.ID <= 0 {
				continue
			}
			setParentPipelineSql += " when ? then ? "

			sqlArgs = append(sqlArgs, base.ID)
			sqlArgs = append(sqlArgs, base.ParentPipelineID)
			setParentPipelineIndex++
		}

		var whereSql = ""
		for index := 0; index < forNum; index++ {
			base := bases[baseIndex]
			if base.ID <= 0 {
				continue
			}
			whereSql += "?,"

			sqlArgs = append(sqlArgs, base.ID)
			baseIndex++
		}
		whereSql = whereSql[:len(whereSql)-1]

		if len(sqlArgs) <= 0 {
			continue
		}

		err := gormClient.Raw(updateSql+setParentTaskSql+"END ,"+setParentPipelineSql+" END "+"where id in ("+whereSql+")", sqlArgs...).Row().Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// this batch insert is set to insert in batches, because mysql batch inserts will have a length limit
func (client *Client) BatchCreatePipelineBases(bases []*spec.PipelineBase, tx *gorm.DB) (err error) {
	if len(bases) <= 0 {
		return nil
	}

	var dbclient = tx
	if tx == nil {
		dbclient = client.DB
	}

	var baseIndex = 0
	var batchInsertNum = 300 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(bases) / batchInsertNum

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(bases) - batchInsertNum*sqlWillRunTimes
		}

		var batchCreatePipelineBases []*spec.PipelineBase
		for index := 0; index < forNum; index++ {
			batchCreatePipelineBases = append(batchCreatePipelineBases, bases[baseIndex])
			baseIndex++
		}
		if err := dbclient.CreateInBatches(&batchCreatePipelineBases, len(batchCreatePipelineBases)).Error; err != nil {
			return err
		}
	}
	return nil
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
