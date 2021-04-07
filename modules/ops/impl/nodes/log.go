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

package nodes

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func (n *Nodes) Logs(req apistructs.OpLogsRequest) (*apistructs.DashboardSpotLogData, error) {
	records, err := n.db.RecordsReader().ByIDs(fmt.Sprintf("%d", req.RecordID)).Do()
	if err != nil {
		errstr := fmt.Sprintf("failed to get record: %v", err)
		logrus.Errorf(errstr)
		return nil, errors.New(errstr)
	}
	if len(records) == 0 {
		errstr := fmt.Sprintf("failed to get record: id: %d", req.RecordID)
		logrus.Errorf(errstr)
		return nil, errors.New(errstr)
	}
	record := records[0]
	task, err := n.bdl.GetPipelineTask(record.PipelineID, req.TaskID)
	if err != nil {
		errstr := fmt.Sprintf("failed to get pipelinetaskinfo: pipelineid: %d, taskid: %d, err: %v", record.PipelineID, req.TaskID, err)
		logrus.Errorf(errstr)
		return nil, errors.New(errstr)
	}
	logData, err := n.bdl.GetLog(apistructs.DashboardSpotLogRequest{
		ID:     task.Extra.UUID,
		Source: apistructs.DashboardSpotLogSourceJob,
		Stream: apistructs.DashboardSpotLogStream(req.Stream),
		Count:  req.Count,
		Start:  req.Start,
		End:    req.End,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to getlog: id: %s, err: %v", task.Extra.UUID, err)
		logrus.Errorf(errstr)
		return nil, errors.New(errstr)
	}
	return logData, nil
}
