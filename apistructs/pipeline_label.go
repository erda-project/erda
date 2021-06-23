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

import (
	"time"
)

// PipelineLabelType defines the type of pipeline label.
type PipelineLabelType string

var (
	PipelineLabelTypeInstance PipelineLabelType = "p_i"   // pipeline instance
	PipelineLabelTypeQueue    PipelineLabelType = "queue" // queue
)

func (t PipelineLabelType) String() string { return string(t) }
func (t PipelineLabelType) Valid() bool {
	switch t {
	case PipelineLabelTypeInstance, PipelineLabelTypeQueue:
		return true
	default:
		return false
	}
}

// TargetIDSelectByLabelRequest select target ids by labels.
type TargetIDSelectByLabelRequest struct {
	Type PipelineLabelType `json:"type"`

	PipelineSources  []PipelineSource `json:"pipelineSource"`
	PipelineYmlNames []string         `json:"pipelineYmlName"`

	// AllowNoMatchLabels, default is false.
	AllowNoMatchLabels bool `json:"allowNoMatchLabels,omitempty"`
	// MUST match
	MustMatchLabels map[string][]string `json:"mustMatchLabels"`
	// ANY match
	AnyMatchLabels map[string][]string `json:"anyMatchLabels"`

	// AllowNoPipelineSources, default is false.
	// 默认查询必须带上 pipeline source，增加区分度
	AllowNoPipelineSources bool `json:"allowNoPipelineSources"`

	// OrderByTargetIDAsc 根据 target_id 升序，默认为 false，即降序
	OrderByTargetIDAsc bool `json:"orderByTargetIDAsc"`
}

type PipelineLabelBatchInsertRequest struct {
	Labels []PipelineLabel `json:"labels"`
}

type PipelineLabelListRequest struct {
	PipelineSource  PipelineSource `schema:"pipelineSource" json:"pipelineSource"`
	PipelineYmlName string         `schema:"pipelineYmlName" json:"pipelineYmlName"`
	TargetIDs       []uint64       `schema:"targetIds" json:"targetIds"`
	MatchKeys       []string       `schema:"matchKeys" json:"matchKeys"`
	UnMatchKeys     []string       `schema:"unMatchKeys" json:"unMatchKeys"`
}

type PipelineLabelPageListData struct {
	Labels []PipelineLabel `json:"labels,omitempty"`
	Total  int64           `json:"total"`
}

type PipelineLabel struct {
	ID              uint64            `json:"id"`
	Type            PipelineLabelType `json:"type"`
	TargetID        uint64            `json:"targetID"`
	PipelineSource  PipelineSource    `json:"pipelineSource"`
	PipelineYmlName string            `json:"pipelineYmlName"`
	Key             string            `json:"key"`
	Value           string            `json:"value"`
	TimeCreated     time.Time         `json:"timeCreated"`
	TimeUpdated     time.Time         `json:"timeUpdated"`
}
