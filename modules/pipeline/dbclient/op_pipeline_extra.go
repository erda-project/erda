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

	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

func (client *Client) GetPipelineExtraByPipelineID(pipelineID uint64, ops ...SessionOption) (spec.PipelineExtra, bool, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var extra spec.PipelineExtra
	extra.PipelineID = pipelineID
	found, err := session.Get(&extra)
	return extra, found, err
}

func (client *Client) CreatePipelineExtra(extra *spec.PipelineExtra, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.InsertOne(extra)
	return err
}

// use insert into on duplicate key update to batch update, pipeline_id absolutely exit, so it always update
func (client *Client) BatchUpdatePipelineExtra(extras []*spec.PipelineExtra, ops ...SessionOption) (err error) {
	if len(extras) <= 0 {
		return nil
	}

	session := client.NewSession(ops...)

	updateSql := "insert into pipeline_extras (pipeline_id, extra) values %s on duplicate key update extra=values(extra)"

	var extraIndex = 0
	var batchInsertNum = 50 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(extras) / batchInsertNum

	var sql []string
	var sqlArgs []interface{}

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(extras) - batchInsertNum*sqlWillRunTimes
		}

		for index := 0; index < forNum; index++ {
			extra := extras[extraIndex]
			sql = append(sql, "(?, ?)")
			sqlArgs = append(sqlArgs, extra.PipelineID)
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.Extra))
			extraIndex++
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
		count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if count <= 0 {
			return fmt.Errorf("insert error: The number of affected rows is %v", count)
		}

		sql = nil
		sqlArgs = nil
	}
	return nil
}

func (client *Client) BatchCreatePipelineExtras(extras []*spec.PipelineExtra, ops ...SessionOption) (err error) {
	if len(extras) <= 0 {
		return nil
	}

	// tx
	session := client.NewSession(ops...)
	insertSql := "INSERT INTO `pipeline_extras`(`pipeline_id`, `pipeline_yml`, `extra`, `normal_labels`, `snapshot`, `commit_detail`, `progress`, `time_created`, " +
		"`time_updated`, `commit`, `org_name`, `snippets`) VALUES %s"

	var extraIndex = 0
	var batchInsertNum = 15 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(extras) / batchInsertNum

	var sql []string
	var sqlArgs []interface{}

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(extras) - batchInsertNum*sqlWillRunTimes
		}

		for index := 0; index < forNum; index++ {
			extra := extras[extraIndex]
			sql = append(sql, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			sqlArgs = append(sqlArgs, extra.PipelineID)
			sqlArgs = append(sqlArgs, extra.PipelineYml)
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.Extra))
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.NormalLabels))
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.Snapshot))
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.CommitDetail))
			sqlArgs = append(sqlArgs, extra.Progress)
			sqlArgs = append(sqlArgs, time.Now())
			sqlArgs = append(sqlArgs, time.Now())
			sqlArgs = append(sqlArgs, extra.Commit)
			sqlArgs = append(sqlArgs, extra.OrgName)
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.Snippets))
			extraIndex++
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
		count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if count <= 0 {
			return fmt.Errorf("insert error: The number of affected rows is %v", count)
		}

		sql = nil
		sqlArgs = nil
	}
	return nil
}

func (client *Client) UpdatePipelineExtraByPipelineID(pipelineID uint64, extra *spec.PipelineExtra, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.Table(&spec.PipelineExtra{}).AllCols().Where("pipeline_id=?", pipelineID).Update(extra)
	return err
}

func (client *Client) UpdatePipelineExtraExtraInfoByPipelineID(pipelineID uint64, extraInfo spec.PipelineExtraInfo, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	// get extra
	extra, found, err := client.GetPipelineExtraByPipelineID(pipelineID, ops...)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf("not found extra")
	}

	// update extra.ExtraInfo
	extra.Extra = extraInfo
	_, err = session.ID(extra.PipelineID).Cols("extra").
		Update(&spec.PipelineExtra{Extra: extraInfo}, spec.PipelineExtra{PipelineID: pipelineID})
	return err
}

func (client *Client) ListPipelineExtrasByPipelineIDs(pipelineIDs []uint64, ops ...SessionOption) (map[uint64]spec.PipelineExtra, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var extras []spec.PipelineExtra
	if err := session.In("pipeline_id", pipelineIDs).Find(&extras); err != nil {
		return nil, err
	}
	extrasMap := make(map[uint64]spec.PipelineExtra, len(extras))
	for _, extra := range extras {
		extrasMap[extra.PipelineID] = extra
	}
	return extrasMap, nil
}

func (client *Client) UpdatePipelineProgress(pipelineID uint64, progress int, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(pipelineID).Cols("progress").Update(&spec.PipelineExtra{Progress: progress})
	return err
}

func (client *Client) UpdatePipelineExtraSnapshot(pipelineID uint64, snapshot spec.Snapshot, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(pipelineID).Cols("snapshot").Update(&spec.PipelineExtra{Snapshot: snapshot})
	return err
}
