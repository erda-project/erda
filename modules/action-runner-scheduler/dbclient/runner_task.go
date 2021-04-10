//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package dbclient

import (
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dbengine"
)

type RunnerTask struct {
	dbengine.BaseModel
	JobID          string `json:"job_id"`
	Status         string `json:"status"` // pending running success failed
	OpenApiToken   string `json:"openapi_token"`
	ContextDataUrl string `json:"context_data_url"`
	ResultDataUrl  string `json:"result_data_url"`
	WorkDir        string `json:"workdir"`
	Commands       string `json:"commands"`
	Targets        string `json:"targets"`
}

// TableName set module's corresponding tableName.
func (RunnerTask) TableName() string {
	return "dice_runner_tasks"
}

func (task RunnerTask) ToApiData() *apistructs.RunnerTask {
	result := &apistructs.RunnerTask{
		ID:             task.ID,
		JobID:          task.JobID,
		Status:         task.Status,
		ContextDataUrl: task.ContextDataUrl,
		OpenApiToken:   task.OpenApiToken,
		ResultDataUrl:  task.ResultDataUrl,
		Commands:       []string{},
		Targets:        []string{},
		WorkDir:        task.WorkDir,
	}
	json.Unmarshal([]byte(task.Commands), &result.Commands)
	json.Unmarshal([]byte(task.Targets), &result.Targets)
	return result
}

func (db *DBClient) CreateRunnerTask(request apistructs.CreateRunnerTaskRequest) (uint64, error) {
	commands, _ := json.Marshal(request.Commands)
	targets, _ := json.Marshal(request.Targets)
	task := &RunnerTask{
		JobID:          request.JobID,
		Status:         apistructs.RunnerTaskStatusPending,
		ContextDataUrl: request.ContextDataUrl,
		ResultDataUrl:  "",
		Commands:       string(commands),
		Targets:        string(targets),
		WorkDir:        request.WorkDir,
	}
	err := db.Save(task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (db *DBClient) GetFirstPendingTask() (*RunnerTask, error) {
	var list []RunnerTask
	err := db.Model(&RunnerTask{}).
		Where("status =?", apistructs.RunnerTaskStatusPending).
		Limit(1).Find(&list).Error
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	result := list[0]
	return &result, nil
}

func (db *DBClient) UpdateRunnerTask(task *RunnerTask) error {
	return db.Save(task).Error
}

func (db *DBClient) GetRunnerTask(id int64) (*RunnerTask, error) {
	var result RunnerTask
	err := db.Model(&RunnerTask{}).Where("id =?", id).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
