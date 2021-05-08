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
	"github.com/pkg/errors"
	"gorm.io/gorm"

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
func (client *Client) BatchUpdatePipelineExtra(extras []*spec.PipelineExtra, tx *gorm.DB) (err error) {
	if len(extras) <= 0 {
		return nil
	}

	var gormClient = tx
	if gormClient == nil {
		gormClient = client.DB
	}
	updateSql := "UPDATE pipeline_extras set "

	var extraIndex = 0
	var batchInsertNum = 50 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(extras) / batchInsertNum

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(extras) - batchInsertNum*sqlWillRunTimes
		}

		var sqlArgs []interface{}

		var setExtraIndex = extraIndex
		var setExtraSql = " extra = case pipeline_id "
		for index := 0; index < forNum; index++ {
			extra := extras[setExtraIndex]
			setExtraSql += " when ? then ? "
			sqlArgs = append(sqlArgs, extra.PipelineID)
			sqlArgs = append(sqlArgs, jsonparse.JsonOneLine(extra.Extra))
			setExtraIndex++
		}

		var whereSql = ""
		for index := 0; index < forNum; index++ {
			extra := extras[extraIndex]
			whereSql += "?,"
			sqlArgs = append(sqlArgs, extra.PipelineID)
			extraIndex++
		}
		whereSql = whereSql[:len(whereSql)-1]

		if len(sqlArgs) <= 0 {
			continue
		}
		err := gormClient.Raw(updateSql+setExtraSql+" END "+"where pipeline_id in ("+whereSql+")", sqlArgs...).Row().Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) BatchCreatePipelineExtras(extras []*spec.PipelineExtra, tx *gorm.DB) (err error) {
	if len(extras) <= 0 {
		return nil
	}

	var gormClient = tx
	if gormClient == nil {
		gormClient = client.DB
	}

	var extraIndex = 0
	var batchInsertNum = 15 // batchInsertNum to limit sql length
	var sqlWillRunTimes = len(extras) / batchInsertNum

	for sqlRunFrequency := 1; sqlRunFrequency <= sqlWillRunTimes+1; sqlRunFrequency++ {

		var forNum = batchInsertNum
		if sqlRunFrequency == sqlWillRunTimes+1 {
			forNum = len(extras) - batchInsertNum*sqlWillRunTimes
		}

		var batchExtras []*spec.PipelineExtra
		for index := 0; index < forNum; index++ {
			batchExtras = append(batchExtras, extras[extraIndex])
			extraIndex++
		}

		if err := gormClient.CreateInBatches(&batchExtras, len(batchExtras)).Error; err != nil {
			return err
		}
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
