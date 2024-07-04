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
	"encoding/json"

	"github.com/erda-project/erda-proto-go/core/pipeline/action_runner_scheduler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type RunnerTask struct {
	dbengine.BaseModel
	JobID          string `json:"job_id"`
	OrgID          int64  `json:"org_id"`
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
		OrgID:          task.OrgID,
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

func (task RunnerTask) ToPbData() *pb.RunnerTask {
	result := &pb.RunnerTask{
		ID:             task.ID,
		JobID:          task.JobID,
		OrgID:          task.OrgID,
		Status:         task.Status,
		ContextDataUrl: task.ContextDataUrl,
		OpenApiToken:   task.OpenApiToken,
		ResultDataUrl:  task.ResultDataUrl,
		Workdir:        task.WorkDir,
		Commands:       []string{},
		Targets:        []string{},
	}
	json.Unmarshal([]byte(task.Commands), &result.Commands)
	json.Unmarshal([]byte(task.Targets), &result.Targets)
	return result
}

func (db *DBClient) CreateRunnerTask(request *pb.RunnerTaskCreateRequest) (uint64, error) {
	commands, _ := json.Marshal(request.Commands)
	targets, _ := json.Marshal(request.Targets)
	task := &RunnerTask{
		JobID:          request.JobID,
		OrgID:          request.OrgID,
		Status:         apistructs.RunnerTaskStatusPending,
		ContextDataUrl: request.ContextDataUrl,
		ResultDataUrl:  "",
		Commands:       string(commands),
		Targets:        string(targets),
		WorkDir:        request.Workdir,
	}
	err := db.Save(task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (db *DBClient) GetFirstPendingTask(orgIDs []int64) (*RunnerTask, error) {
	var list []RunnerTask
	sql := db.Model(&RunnerTask{}).
		Where("status = ?", apistructs.RunnerTaskStatusPending)
	// handle org id
	if len(orgIDs) > 0 {
		var polishedOrgIDs []int64
		for _, orgID := range orgIDs {
			if orgID > 0 {
				polishedOrgIDs = append(polishedOrgIDs, orgID)
			}
		}
		if len(polishedOrgIDs) > 0 {
			sql = sql.Where("org_id in (?)", orgIDs)
		}
	}
	// order, oldest first
	sql = sql.Order("id ASC")
	err := sql.Limit(1).Find(&list).Error
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
