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
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/base/converter"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/pbutil"
)

// Rerun commit 不变
func (s *PipelineSvc) Rerun(req *basepb.PipelineRerunRequest) (*spec.Pipeline, error) {

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

	if origin.Labels == nil {
		origin.Labels = map[string]string{}
	}
	origin.Labels[apistructs.LabelPipelineType] = apistructs.PipelineTypeRerun.String()

	p, err := s.CreateV2(&basepb.PipelineCreateRequest{
		PipelineYml:            origin.PipelineYml,
		ClusterName:            origin.ClusterName,
		PipelineYmlName:        origin.PipelineYmlName,
		RunParams:              converter.ToPipelineRunParams(origin.Snapshot.RunPipelineParams),
		PipelineSource:         origin.PipelineSource.String(),
		Labels:                 origin.Labels,
		NormalLabels:           origin.GenerateNormalLabelsForCreateV2(),
		Envs:                   origin.Snapshot.Envs,
		ConfigManageNamespaces: origin.GetConfigManageNamespaces(),
		AutoRunAtOnce:          req.AutoRunAtOnce,
		AutoStartCron:          false,
		CronStartFrom:          pbutil.GetTimestamp(originCron.Extra.CronStartFrom),
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
