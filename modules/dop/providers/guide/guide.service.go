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

package guide

import (
	"context"
	"sort"

	"github.com/erda-project/erda-proto-go/dop/guide/pb"
	projectPipelinePb "github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/guide/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

type GuideService struct {
	bdl *bundle.Bundle
	db  *db.GuideDB

	ProjectPipelineService projectPipelinePb.ProjectPipelineServiceServer
}

func (g *GuideService) ListGuide(ctx context.Context, req *pb.ListGuideRequest) (*pb.ListGuideResponse, error) {
	guidesDB, err := g.db.ListGuide(req)
	if err != nil {
		return nil, apierrors.ErrListGuide.InvalidParameter(err)
	}
	guides := make([]*pb.Guide, 0, len(guidesDB))
	for _, v := range guidesDB {
		guides = append(guides, v.Convert())
	}
	return &pb.ListGuideResponse{Data: guides}, nil
}

func (g *GuideService) JudgeCanCreatePipeline(ctx context.Context, req *pb.JudgeCanCreatePipelineRequest) (*pb.JudgeCanCreatePipelineResponse, error) {
	guideDB, err := g.db.GetGuide(req.ID)
	if err != nil {
		return nil, apierrors.ErrJudgeCanCreatePipeline.InvalidParameter(err)
	}

	ymlResp, err := g.ProjectPipelineService.ListPipelineYml(ctx, &projectPipelinePb.ListAppPipelineYmlRequest{
		AppID:  guideDB.AppID,
		Branch: guideDB.Branch,
	})
	if err != nil {
		return nil, apierrors.ErrJudgeCanCreatePipeline.InvalidParameter(err)
	}

	var (
		path      string
		fileName  string
		canCreate bool
	)

	// pipeline.yml first
	sort.Slice(ymlResp.Result, func(i, j int) bool {
		return ymlResp.Result[i].YmlPath < ymlResp.Result[j].YmlPath
	})
	if len(ymlResp.Result) > 0 {
		canCreate = true
		path = ymlResp.Result[0].YmlPath
		fileName = ymlResp.Result[0].YmlName
	}

	return &pb.JudgeCanCreatePipelineResponse{
		CanCreate: canCreate,
		AppID:     guideDB.AppID,
		Branch:    guideDB.Branch,
		Path:      path,
		FileName:  fileName,
	}, nil
}
