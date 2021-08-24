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

package spec

import (
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
)

const (
	//pipelineCron表的字段名
	PipelineCronCronExpr = "cron_expr"
	PipelineCronEnable   = "enable"
	Extra                = "extra"
)

type PipelineCron struct {
	ID          uint64    `json:"id" xorm:"pk autoincr"`
	TimeCreated time.Time `json:"timeCreated" xorm:"created"` // 记录创建时间
	TimeUpdated time.Time `json:"timeUpdated" xorm:"updated"` // 记录更新时间

	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`

	CronExpr string `json:"cronExpr"`
	//PipelineSource  string            `json:"pipelineSource"`
	Enable *bool             `json:"enable"` // 1 true, 0 false
	Extra  PipelineCronExtra `json:"extra,omitempty" xorm:"json"`

	// Deprecated
	ApplicationID uint64 `json:"applicationID" xorm:"application_id"`
	// Deprecated
	Branch string `json:"branch"`
	// Deprecated
	BasePipelineID uint64 `json:"basePipelineID"` // 用于记录最开始创建出这条 cron 记录的 pipeline id
}

// PipelineCronExtra cron 扩展信息, 不参与过滤
type PipelineCronExtra struct {
	PipelineYml            string            `json:"pipelineYml"`
	ClusterName            string            `json:"clusterName"`
	FilterLabels           map[string]string `json:"labels"`
	NormalLabels           map[string]string `json:"normalLabels"` // userID 存储提交流水线的用户 ID
	Envs                   map[string]string `json:"envs"`
	ConfigManageNamespaces []string          `json:"configManageNamespaces,omitempty"`
	CronStartFrom          *time.Time        `json:"cronStartFrom,omitempty"`
	// 新版为 v2
	Version string `json:"version"`

	// compensate
	// Compensator 老的 cron 为空，经过补偿后会自动赋值默认配置；新创建的 cron 一定会有值。
	Compensator *apistructs.CronCompensator `json:"compensator,omitempty"`
	//每次中断补偿执行的时间，下次中断补偿从这个时间开始查询
	LastCompensateAt *time.Time `json:"lastCompensateAt,omitempty"`
}

func (PipelineCron) TableName() string {
	return "pipeline_crons"
}

func (pc *PipelineCron) Convert2DTO() *apistructs.PipelineCronDTO {
	if pc == nil {
		return nil
	}
	return &apistructs.PipelineCronDTO{
		ID:              pc.ID,
		TimeCreated:     pc.TimeCreated,
		TimeUpdated:     pc.TimeUpdated,
		ApplicationID:   pc.ApplicationID,
		Branch:          pc.Branch,
		CronExpr:        pc.CronExpr,
		CronStartTime:   pc.Extra.CronStartFrom,
		PipelineYmlName: pc.PipelineYmlName,
		BasePipelineID:  pc.BasePipelineID,
		Enable:          pc.Enable,
	}
}

// GetAppID 返回 AppID，若为 0 则表示不存在
// 优先级如下：
// 1. pc.AppID
// 2. pc.Extra.Labels["AppID"]
func (pc *PipelineCron) GetAppID() uint64 {
	if pc == nil {
		return 0
	}
	if pc.ApplicationID > 0 {
		return pc.ApplicationID
	}
	if len(pc.Extra.FilterLabels) == 0 {
		return 0
	}
	// ignore parse err
	appID, _ := strconv.ParseUint(pc.Extra.FilterLabels[apistructs.LabelAppID], 10, 64)
	return appID
}

// GetBranch 返回 Branch, "" 则表示不存在
// 优先级如下：
// 1. pc.Branch
// 2. pc.Extra.Labels["Branch"]
func (pc *PipelineCron) GetBranch() string {
	if pc == nil {
		return ""
	}
	if pc.Branch != "" {
		return pc.Branch
	}
	if len(pc.Extra.FilterLabels) == 0 {
		return ""
	}
	return pc.Extra.FilterLabels[apistructs.LabelBranch]
}
