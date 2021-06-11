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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) BatchCreateLabels(createReq *apistructs.PipelineLabelBatchInsertRequest) error {
	var insertLabels []spec.PipelineLabel
	for _, label := range createReq.Labels {
		insertLabels = append(insertLabels, spec.PipelineLabel{
			Type:            label.Type,
			Key:             label.Key,
			Value:           label.Value,
			TargetID:        label.TargetID,
			PipelineYmlName: label.PipelineYmlName,
			PipelineSource:  label.PipelineSource,
			TimeCreated:     time.Now(),
			TimeUpdated:     time.Now(),
		})
	}
	err := s.dbClient.BatchInsertLabels(insertLabels)
	if err != nil {
		return err
	}
	return nil
}

func (s *PipelineSvc) ListLabels(req *apistructs.PipelineLabelListRequest) (*apistructs.PipelineLabelPageListData, error) {

	// if UnMatchKeys and TargetIDS not empty, find targetID by MatchKeys from db then filter targetID not in TargetIDs
	if len(req.UnMatchKeys) > 0 && len(req.TargetIDs) > 0 {
		req.MatchKeys = req.UnMatchKeys
	}

	labels, _, err := s.dbClient.ListPipelineLabels(req)
	if err != nil {
		return nil, apierrors.ErrListPipelineLabel.InternalError(err)
	}

	var listLabel []apistructs.PipelineLabel
	// filter targetID not in TargetIDS
	if len(req.UnMatchKeys) > 0 && len(req.TargetIDs) > 0 {
		var labelMap = make(map[uint64]spec.PipelineLabel)
		for _, value := range labels {
			labelMap[value.TargetID] = value
		}

		for _, value := range req.TargetIDs {
			if _, ok := labelMap[value]; ok {
				continue
			}

			listLabel = append(listLabel, apistructs.PipelineLabel{
				TargetID: value,
			})
		}
	} else {
		for _, value := range labels {
			listLabel = append(listLabel, apistructs.PipelineLabel{
				ID:              value.ID,
				Type:            value.Type,
				TargetID:        value.TargetID,
				PipelineSource:  value.PipelineSource,
				PipelineYmlName: value.PipelineYmlName,
				Key:             value.Key,
				Value:           value.Value,
				TimeCreated:     value.TimeCreated,
				TimeUpdated:     value.TimeUpdated,
			})
		}
	}

	var result apistructs.PipelineLabelPageListData
	result.Labels = listLabel
	return &result, nil
}
