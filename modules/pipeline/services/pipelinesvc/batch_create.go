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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *PipelineSvc) BatchCreate(batchReq *apistructs.PipelineBatchCreateRequest) (
	map[string]*apistructs.PipelineDTO, error) {

	// convert pipelineBatchCreateRequest to []pipelineCreateRequest
	var reqs []*apistructs.PipelineCreateRequest
	for _, ymlPath := range batchReq.BatchPipelineYmlPaths {
		reqs = append(reqs, &apistructs.PipelineCreateRequest{
			AppID:             batchReq.AppID,
			Branch:            batchReq.Branch,
			Source:            batchReq.Source,
			PipelineYmlName:   ymlPath,
			CallbackURLs:      batchReq.CallbackURLs,
			AutoRun:           batchReq.AutoRun,
			PipelineYmlSource: apistructs.PipelineYmlSourceGittar, // 通过批量接口创建暂时只需要支持 gittar 类型
			UserID:            batchReq.UserID,
		})
	}

	result := make(map[string]*apistructs.PipelineDTO)

	for _, req := range reqs {
		p, err := s.makePipelineFromRequest(req)
		if err != nil {
			return nil, apierrors.ErrBatchCreatePipeline.InternalError(err)
		}
		if err = s.createPipelineGraph(p); err != nil {
			return nil, apierrors.ErrBatchCreatePipeline.InternalError(err)
		}

		identityInfo := apistructs.IdentityInfo{UserID: req.UserID}

		// 是否自动执行
		if batchReq.AutoRun {
			if p, err = s.RunPipeline(&apistructs.PipelineRunRequest{
				PipelineID:   p.ID,
				IdentityInfo: identityInfo},
			); err != nil {
				return nil, apierrors.ErrRunPipeline.InternalError(err)
			}
		}
		result[req.PipelineYmlName] = s.ConvertPipeline(p)
	}

	return result, nil
}
