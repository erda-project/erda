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
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/retry"
)

func (client *Client) CreatePipelineStage(ps *spec.PipelineStage, ops ...SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err = session.InsertOne(ps)
	return err
}

// this batch insert is set to insert in batches, because mysql batch inserts will have a length limit
func (client *Client) BatchCreatePipelineStages(stages []*spec.PipelineStage, ops ...SessionOption) (err error) {
	if len(stages) <= 0 {
		return nil
	}

	// tx
	session := client.NewSession(ops...)

	insertSql := "INSERT INTO `pipeline_stages`(`pipeline_id`, `name`, `extra`, `status`, `cost_time_sec`, " +
		"`time_begin`, `time_end`, `time_created`, `time_updated`) VALUES %s"

	var stageIndex = 0
	var batchInsertNum = 300 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(stages) / batchInsertNum

	var sql []string
	var sqlArgs []interface{}

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(stages) - batchInsertNum*sqlWillRunTimes
		}

		for index := 0; index < forNum; index++ {
			stage := stages[stageIndex]
			sql = append(sql, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			sqlArgs = append(sqlArgs, stage.PipelineID)
			sqlArgs = append(sqlArgs, stage.Name)
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(stage.Extra))
			sqlArgs = append(sqlArgs, stage.Status.String())
			sqlArgs = append(sqlArgs, stage.CostTimeSec)
			sqlArgs = append(sqlArgs, stage.TimeBegin)
			sqlArgs = append(sqlArgs, stage.TimeEnd)
			sqlArgs = append(sqlArgs, time.Now())
			sqlArgs = append(sqlArgs, time.Now())
			stageIndex++
		}

		if len(sqlArgs) <= 0 {
			continue
		}

		stmt := fmt.Sprintf(insertSql, strings.Join(sql, ","))
		sqlExec := append([]interface{}{stmt}, sqlArgs...)
		res, err := session.Exec(sqlExec...)
		if err != nil {
			return err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// decreasingly set the id of the task
		for i := 0; i < forNum; i++ {
			stages[stageIndex-forNum+i].ID = uint64(lastID)
			lastID++
		}

		sql = nil
		sqlArgs = nil
	}

	return nil
}

func (client *Client) GetPipelineStage(id interface{}, ops ...SessionOption) (spec.PipelineStage, error) {
	session := client.NewSession(ops...)
	defer session.Close()
	var stage spec.PipelineStage
	exist, err := session.ID(id).Get(&stage)
	if err != nil {
		return spec.PipelineStage{}, errors.Wrapf(err, "failed to get stage by id [%v]", id)
	}
	if !exist {
		return spec.PipelineStage{}, errors.Errorf("not found stage by id [%v]", id)
	}
	return stage, nil
}

func (client *Client) GetPipelineStageWithPreStatus(id interface{}, ops ...SessionOption) (spec.PipelineStage, error) {
	session := client.NewSession(ops...)
	defer session.Close()
	stage, err := client.GetPipelineStage(id, ops...)
	if err != nil {
		return spec.PipelineStage{}, err
	}
	if stage.Extra.PreStage != nil && stage.Extra.PreStage.ID > 0 {
		preStage, err := client.GetPipelineStage(stage.Extra.PreStage.ID, ops...)
		if err != nil {
			return spec.PipelineStage{}, errors.Wrap(err, "failed to get pre stage")
		}
		stage.Extra.PreStage.Status = preStage.Status
	}
	return stage, nil
}

func (client *Client) UpdatePipelineStage(id interface{}, stage *spec.PipelineStage) error {
	if _, err := client.ID(id).AllCols().Update(stage); err != nil {
		return errors.Wrapf(err, "failed to update stage, id [%v]", id)
	}
	return nil
}

func (client *Client) ListPipelineStageByPipelineID(pipelineID uint64, ops ...SessionOption) ([]spec.PipelineStage, error) {
	session := client.NewSession(ops...)
	defer session.Close()
	var stageList []spec.PipelineStage
	if err := session.Find(&stageList, spec.PipelineStage{PipelineID: pipelineID}); err != nil {
		return nil, err
	}
	for i, stage := range stageList {
		tmp, err := client.GetPipelineStageWithPreStatus(stage.ID, ops...)
		if err != nil {
			return nil, err
		}
		stageList[i] = tmp
	}
	return stageList, nil
}

func (client *Client) ListPipelineStageByStatuses(statuses ...apistructs.PipelineStatus) ([]spec.PipelineStage, error) {
	var stageList []spec.PipelineStage
	if err := client.In("status", statuses).Find(&stageList); err != nil {
		return nil, err
	}
	for i, stage := range stageList {
		tmp, err := client.GetPipelineStageWithPreStatus(stage.ID)
		if err != nil {
			return nil, err
		}
		stageList[i] = tmp
	}
	return stageList, nil
}

func (client *Client) DeletePipelineStagesByPipelineID(pipelineID uint64, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	return retry.DoWithInterval(func() error {
		_, err := session.Delete(&spec.PipelineStage{PipelineID: pipelineID})
		return err
	}, 3, time.Second)
}
