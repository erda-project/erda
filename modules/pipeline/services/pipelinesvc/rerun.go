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
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// Rerun commit 不变
func (s *PipelineSvc) Rerun(req *apistructs.PipelineRerunRequest) (*spec.Pipeline, error) {

	origin, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRerunPipeline.InternalError(err)
	}

	var originCron spec.PipelineCron
	if origin.CronID != nil {
		cron, err := s.dbClient.GetPipelineCron(*origin.CronID)
		if err != nil {
			return nil, apierrors.ErrRerunPipeline.InternalError(err)
		}
		originCron = cron
	}

	p, err := s.CreateV2(&apistructs.PipelineCreateRequestV2{
		PipelineYml:            origin.PipelineYml,
		ClusterName:            origin.ClusterName,
		PipelineYmlName:        origin.PipelineYmlName,
		RunParams:              origin.Snapshot.RunPipelineParams.ToPipelineRunParams(),
		PipelineSource:         origin.PipelineSource,
		Labels:                 origin.Labels,
		NormalLabels:           origin.GenerateNormalLabelsForCreateV2(),
		Envs:                   origin.Snapshot.Envs,
		ConfigManageNamespaces: origin.GetConfigManageNamespaces(),
		AutoRunAtOnce:          req.AutoRunAtOnce,
		AutoStartCron:          false,
		CronStartFrom:          originCron.Extra.CronStartFrom,
		IdentityInfo:           req.IdentityInfo,
	})
	if err != nil {
		return nil, err
	}

	p.Extra.CopyFromPipelineID = &origin.ID

	if err := s.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(p.ID, p.Extra); err != nil {
		return nil, apierrors.ErrUpdatePipeline.InternalError(err)
	}

	return p, nil
}
