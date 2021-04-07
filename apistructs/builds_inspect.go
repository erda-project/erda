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

// import (
// 	"time"
// )
//
// type TriggerMode string
//
// var (
// 	TriggerModeOnce TriggerMode = "once" // 单次触发
// 	TriggerModeCron TriggerMode = "cron" // 定时触发
// )
//
// type (
// 	Version map[string]string
//
// 	MetadataField struct {
// 		Name     string `json:"name"`
// 		Value    string `json:"value"`
// 		Type     string `json:"type,omitempty"`
// 		Optional bool   `json:"optional,omitempty"`
// 	}
//
// 	Metadata []MetadataField
//
// 	Extra struct {
// 		Version  Version  `json:"version"`
// 		Metadata Metadata `json:"metadata"`
// 		Errors   []Error  `json:"errors"`
// 		// TODO add storage info? OutStorage/InStorage
// 	}
// )
//
// type Error struct {
// 	Code    string                 `json:"code,omitempty"`
// 	Message string                 `json:"msg,omitempty"`
// 	Ctx     map[string]interface{} `json:"ctx,omitempty"`
// }
//
// type BuildResponseDTO struct {
// 	Success bool       `json:"success"`
// 	Data    BuildForUI `json:"data,omitempty"`
// 	Error   BuildError `json:"err,omitempty"`
// }
//
// type BuildForUI struct {
// 	Build     CiV3Builds                      `json:"build"`
// 	Envs      map[string]string               `json:"envs,omitempty"`
// 	Stages    []StageForUI                    `json:"stages"`
// 	Instances []PipelineInstanceExecuteRecord `json:"instances,omitempty"`
//
// 	// 触发方式
// 	ParsedTriggerMode TriggerMode `json:"parsed_trigger_mode"` // 从 pipeline.yml 解析出来的执行方式
//
// 	// 按钮
// 	OnceRunnable  bool `json:"once_runnable"`  // 是否可单次执行
// 	Rerunnable    bool `json:"rerunnable"`     // 是否可重试
// 	CronStartable bool `json:"cron_startable"` // cron 是否可开始
// 	CronStoppable bool `json:"cron_stoppable"` // cron 是否可停止
//
// 	// cron
// 	CronExpr string `json:"cron_time,omitempty"`
// }
//
// type StageForUI struct {
// 	ID     string      `json:"id,omitempty"`
// 	Name   string      `json:"name,omitempty"`
// 	Uuid   string      `json:"uuid,omitempty"`
// 	Status string      `json:"status,omitempty"`
// 	Tasks  []TaskForUI `json:"tasks,omitempty"`
// }
//
// type TaskForUI struct {
// 	ID             string            `json:"id,omitempty"`
// 	Name           string            `json:"name,omitempty"`
// 	Uuid           string            `json:"uuid,omitempty"`
// 	Type           string            `json:"type"`
// 	Status         string            `json:"status,omitempty"`
// 	Envs           map[string]string `json:"envs,omitempty"`
// 	BpArgs         map[string]string `json:"bp_args,omitempty"`
// 	BpRepoArgs     map[string]string `json:"bp_repo_args,omitempty"`
// 	ForceBuildpack bool              `json:"force_buildpack"`
// 	Disabled       bool              `json:"disabled"`
// 	Time           int               `json:"time"`
// 	StageName      string            `json:"stageName,omitempty"`
// 	LogId          string            `json:"logId"`
// 	Extra      *Extra        `json:"extra_info,omitempty"`
// }
//
// type PipelineInstanceExecuteRecord struct {
// 	RerunVer  *int64     `json:"rerun_ver"`          // 执行版本
// 	Status    *string    `json:"status"`             // 实例状态
// 	TimeBegin *time.Time `json:"time_begin"`         // 真正开始执行的时间
// 	TimeEnd   *time.Time `json:"time_end,omitempty"` // 结束时间
// }
