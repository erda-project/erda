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
	"fmt"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) DeletePipelineRelated(pipelineID uint64, ops ...SessionOption) error {
	// 获取 pipeline
	p, err := client.GetPipeline(pipelineID, ops...)
	if err != nil {
		return err
	}
	// 校验当前流水线是否可被删除
	can, reason := canDelete(p)
	if !can {
		return fmt.Errorf("cannot delete, reason: %s", reason)
	}

	// pipelines
	if err := client.DeletePipeline(pipelineID, ops...); err != nil {
		return err
	}

	// related pipeline stages
	if err := client.DeletePipelineStagesByPipelineID(pipelineID, ops...); err != nil {
		return err
	}

	// related pipeline tasks
	if err := client.DeletePipelineTasksByPipelineID(pipelineID, ops...); err != nil {
		return err
	}

	// related pipeline labels
	if err := client.DeletePipelineLabelsByPipelineID(pipelineID, ops...); err != nil {
		return err
	}

	// related pipeline reports
	if err := client.DeletePipelineReportsByPipelineID(pipelineID, ops...); err != nil {
		return err
	}

	return nil
}

func canDelete(p spec.Pipeline) (bool, string) {
	// status
	if !p.Status.CanDelete() {
		return false, fmt.Sprintf("invalid status: %s", p.Status)
	}
	// 终态后需要判断 complete gc
	if p.Status.IsEndStatus() {
		if !p.Extra.CompleteReconcilerGC {
			return false, fmt.Sprintf("waiting gc")
		}
	}
	return true, ""
}

func canArchive(p spec.Pipeline) (bool, string) {
	return canDelete(p)
}
