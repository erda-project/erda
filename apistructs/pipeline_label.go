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

type PipelineIDSelectByLabelRequest struct {
	PipelineSources  []PipelineSource `json:"pipelineSource"`
	PipelineYmlNames []string         `json:"pipelineYmlName"`

	// MUST match
	MustMatchLabels map[string][]string `json:"mustMatchLabels"`
	// ANY match
	AnyMatchLabels map[string][]string `json:"anyMatchLabels"`

	// AllowNoPipelineSources, default is false.
	// 默认查询必须带上 pipeline source，增加区分度
	AllowNoPipelineSources bool `json:"allowNoPipelineSources"`

	// OrderByPipelineIDASC 根据 pipeline_id 升序，默认为 false，即降序
	OrderByPipelineIDASC bool `json:"orderByPipelineIDDesc"`
}
