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

package label

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/pipeline/label/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type labelServiceImpl struct {
	dbClient *dbclient.Client
}

func (l *labelServiceImpl) BatchCreateLabels(createReq *pb.PipelineLabelBatchInsertRequest) error {
	var insertLabels []spec.PipelineLabel
	for _, label := range createReq.Labels {
		insertLabels = append(insertLabels, spec.PipelineLabel{
			Type:            apistructs.PipelineLabelType(label.Type),
			Key:             label.Key,
			Value:           label.Value,
			TargetID:        label.TargetID,
			PipelineYmlName: label.PipelineYmlName,
			PipelineSource:  apistructs.PipelineSource(label.PipelineSource),
			TimeCreated:     time.Now(),
			TimeUpdated:     time.Now(),
			ID:              uuid.SnowFlakeIDUint64(),
		})
	}
	err := l.dbClient.BatchInsertLabels(insertLabels)
	if err != nil {
		return err
	}
	return nil
}

func (l *labelServiceImpl) ListLabels(req *pb.PipelineLabelListRequest) (*pb.PipelineLabelPageListData, error) {
	// if UnMatchKeys and TargetIds not empty, find targetID by MatchKeys from db then filter targetID not in TargetIds
	if len(req.UnMatchKeys) > 0 && len(req.TargetIDs) > 0 {
		req.MatchKeys = req.UnMatchKeys
	}

	labels, _, err := l.dbClient.ListPipelineLabels(req)
	if err != nil {
		return nil, apierrors.ErrListPipelineLabel.InternalError(err)
	}

	var listLabel []*pb.PipelineLabel
	// filter targetID not in TargetIds
	if len(req.UnMatchKeys) > 0 && len(req.TargetIDs) > 0 {
		var labelMap = make(map[uint64]spec.PipelineLabel)
		for _, value := range labels {
			labelMap[value.TargetID] = value
		}

		for _, value := range req.TargetIDs {
			if _, ok := labelMap[value]; ok {
				continue
			}

			listLabel = append(listLabel, &pb.PipelineLabel{
				TargetID: value,
			})
		}
	} else {
		for _, value := range labels {
			listLabel = append(listLabel, &pb.PipelineLabel{
				ID:              value.ID,
				Type:            value.Type.String(),
				TargetID:        value.TargetID,
				PipelineSource:  value.PipelineSource.String(),
				PipelineYmlName: value.PipelineYmlName,
				Key:             value.Key,
				Value:           value.Value,
				TimeCreated:     timestamppb.New(value.TimeCreated),
				TimeUpdated:     timestamppb.New(value.TimeUpdated),
			})
		}
	}

	var result pb.PipelineLabelPageListData
	result.Labels = listLabel
	return &result, nil
}
