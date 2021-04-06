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

package apistructs

const (
	RunnerTaskStatusPending  = "pending"
	RunnerTaskStatusRunning  = "running"
	RunnerTaskStatusSuccess  = "success"
	RunnerTaskStatusFailed   = "failed"
	RunnerTaskStatusCanceled = "canceled"
)

type RunnerTask struct {
	ID             uint64   `json:"id"`
	JobID          string   `json:"job_id"`
	Status         string   `json:"status"` // pending running success failed
	ContextDataUrl string   `json:"context_data_url"`
	OpenApiToken   string   `json:"openapi_token"`
	ResultDataUrl  string   `json:"result_data_url"`
	Commands       []string `json:"commands"`
	Targets        []string `json:"targets"`
	WorkDir        string   `json:"workdir"`
}

type QueryRunnerTaskRequest struct {
	TaskID string
}

type QueryRunnerTaskResponse struct {
	Header
	Data RunnerTask `json:"data"`
}

type CreateRunnerTaskRequest struct {
	JobID          string   `json:"job_id"`
	ContextDataUrl string   `json:"context_data_url"`
	Commands       []string `json:"commands"`
	Targets        []string `json:"targets"`
	WorkDir        string   `json:"workdir"`
}

type CreateRunnerTaskResponse struct {
	Header
	Data int64 `json:"data"`
}

type UpdateRunnerTaskRequest struct {
	ID             int64  `json:"-"`
	TaskID         string `json:"task_id"`
	Status         string `json:"status"`
	ContextDataUrl string `json:"context_data_url"`
	ResultDataUrl  string `json:"result_data_url"`
}
