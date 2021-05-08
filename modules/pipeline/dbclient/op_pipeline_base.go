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
func (client *Client) BatchUpdatePipelineBaseParentID(bases []*spec.PipelineBase, ops ...SessionOption) (err error) {
	if len(bases) <= 0 {
		return nil
	}

	// tx
	session := client.NewSession(ops...)

	updateSql := "insert into pipeline_bases (id, parent_task_id, parent_pipeline_id, is_snippet) values %s " +
		"on duplicate key update parent_task_id=values(parent_task_id), parent_pipeline_id=values(parent_pipeline_id), is_snippet=values(is_snippet);"

	var baseIndex = 0
	var batchInsertNum = 500 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(bases) / batchInsertNum

	var sql []string
	var sqlArgs []interface{}

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(bases) - batchInsertNum*sqlWillRunTimes
		}

		for index := 0; index < forNum; index++ {
			base := bases[baseIndex]
			sql = append(sql, "(?, ?, ?, ?)")
			sqlArgs = append(sqlArgs, base.ID)
			sqlArgs = append(sqlArgs, base.ParentTaskID)
			sqlArgs = append(sqlArgs, base.ParentPipelineID)
			sqlArgs = append(sqlArgs, 1)
			baseIndex++
		}

		if len(sqlArgs) <= 0 {
			continue
		}

		stmt := fmt.Sprintf(updateSql, strings.Join(sql, ","))
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
			bases[baseIndex-forNum+i].ID = uint64(lastID)
			lastID++
		}

		sql = nil
		sqlArgs = nil
	}
	return nil
}

// this batch insert is set to insert in batches, because mysql batch inserts will have a length limit
func (client *Client) BatchCreatePipelineBases(bases []*spec.PipelineBase, ops ...SessionOption) (err error) {
	if len(bases) <= 0 {
		return nil
	}

	// tx
	session := client.NewSession(ops...)

	insertSql := "INSERT INTO `pipeline_bases`(`pipeline_source`, `pipeline_yml_name`, `cluster_name`, " +
		"`status`, `type`, `trigger_mode`, `cron_id`, `is_snippet`, `parent_pipeline_id`, `parent_task_id`, `cost_time_sec`, " +
		"`time_begin`, `time_end`, `time_created`, `time_updated`) VALUES %s"

	var baseIndex = 0
	var batchInsertNum = 300 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(bases) / batchInsertNum

	var sql []string
	var sqlArgs []interface{}

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(bases) - batchInsertNum*sqlWillRunTimes
		}

		for index := 0; index < forNum; index++ {
			base := bases[baseIndex]
			sql = append(sql, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			sqlArgs = append(sqlArgs, base.PipelineSource)
			sqlArgs = append(sqlArgs, base.PipelineYmlName)
			sqlArgs = append(sqlArgs, base.ClusterName)
			sqlArgs = append(sqlArgs, base.Status)
			sqlArgs = append(sqlArgs, base.Type.String())
			sqlArgs = append(sqlArgs, base.TriggerMode)
			sqlArgs = append(sqlArgs, base.CronID)
			sqlArgs = append(sqlArgs, base.IsSnippet)
			sqlArgs = append(sqlArgs, base.ParentPipelineID)
			sqlArgs = append(sqlArgs, base.ParentTaskID)
			sqlArgs = append(sqlArgs, base.CostTimeSec)
			sqlArgs = append(sqlArgs, base.TimeBegin)
			sqlArgs = append(sqlArgs, base.TimeEnd)
			sqlArgs = append(sqlArgs, time.Now())
			sqlArgs = append(sqlArgs, time.Now())
			baseIndex++
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
			bases[baseIndex-forNum+i].ID = uint64(lastID)
			lastID++
		}

		sql = nil
		sqlArgs = nil
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
