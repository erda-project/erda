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

import "time"

type PipelineBasicReport struct {
	PipelineSource   PipelineSource `json:"pipelineSource"`
	PipelineYmlName  string         `json:"pipelineYmlName"`
	ClusterName      string         `json:"clusterName"`
	TimeCreated      *time.Time     `json:"timeCreated,omitempty"`
	TimeBegin        *time.Time     `json:"timeBegin,omitempty"`
	TimeEnd          *time.Time     `json:"timeEnd,omitempty"`
	TotalCostTimeSec int64          `json:"totalCostTimeSec"`

	TaskInfos []TaskReportInfo `json:"taskInfos,omitempty"`
}

type TaskReportInfo struct {
	Name             string                   `json:"name"`
	ActionType       string                   `json:"actionType"`
	ActionVersion    string                   `json:"actionVersion"`
	ExecutorType     string                   `json:"executorType"`
	ClusterName      string                   `json:"clusterName"`
	TimeBegin        *time.Time               `json:"timeBegin,omitempty"`
	TimeEnd          *time.Time               `json:"timeEnd,omitempty"`
	TimeBeginQueue   *time.Time               `json:"timeBeginQueue,omitempty"`
	TimeEndQueue     *time.Time               `json:"timeEndQueue,omitempty"`
	QueueCostTimeSec int64                    `json:"queueCostTimeSec"`
	RunCostTimeSec   int64                    `json:"runCostTimeSec"`
	MachineStat      *PipelineTaskMachineStat `json:"machineStat,omitempty"`
	Meta             map[string]string        `json:"meta"`
}
