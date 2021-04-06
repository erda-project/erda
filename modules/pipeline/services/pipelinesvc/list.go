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

package pipelinesvc

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) List(condition apistructs.PipelinePageListRequest) (*apistructs.PipelinePageListData, error) {
	pipelines, _, total, currentPageSize, err := s.dbClient.PageListPipelines(condition)
	if err != nil {
		return nil, apierrors.ErrListPipeline.InternalError(err)
	}

	var result apistructs.PipelinePageListData
	result.Pipelines = s.BatchConvert2PagePipeline(pipelines)
	result.Total = total
	result.CurrentPageSize = currentPageSize

	return &result, nil
}

// calculateProgress 根据 pipeline tasks 计算 progress
// 若 progress < 0，认为进度还没最终确认，包括 running 或 还没计算过 的情况；
// 若 progress >= 0，认为进度已经计算完毕（包括 0），直接返回
func (s *PipelineSvc) calculateProgress(p spec.Pipeline) (int, error) {

	if p.Progress >= 0 {
		// pipeline 为成功状态，progress 不应该为 0，需要重新计算
		if p.Progress == 0 && p.Status.IsSuccessStatus() {
			// calculate progress
			goto CalculateStatus
		}
		// progress >= 0，直接返回
		return p.Progress, nil
	}

CalculateStatus:
	needStoreToDB := false
	if p.Status.IsEndStatus() {
		needStoreToDB = true
	}

	// calculate pipeline progress
	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return -1, err
	}
	var successCount int
	for _, t := range tasks {
		if t.Status.IsSuccessStatus() {
			successCount++
		}
	}
	var progress int
	if len(tasks) == 0 { // 存在 task 为 0 的情况
		progress = 100
	} else {
		progress = (successCount / len(tasks)) * 100
	}

	if needStoreToDB {
		go func() {
			// 异步更新 pipeline progress
			err := s.dbClient.UpdatePipelineProgress(p.ID, progress)
			if err != nil {
				logrus.Errorf("[alert] failed to update pipeline progress, pipelineID: %d, progress: %d, err: %v",
					p.ID, progress, err)
			}
		}()
	}
	return progress, nil
}
