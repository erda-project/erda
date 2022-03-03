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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/dop/guide/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/guide/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/limit_sync_group"
)

const (
	DicePipelinePath string = ".dice/pipelines"
	ErdaPipelinePath string = ".erda/pipelines"
	InitCommitID     string = "0000000000000000000000000000000000000000"
	BranchPrefix     string = "refs/heads/"
)

type GuideService struct {
	bdl *bundle.Bundle
	db  *db.GuideDB
}

func (g *GuideService) CreateGuideByGittarHook(ctx context.Context, req *pb.GittarPushPayloadEvent) (*pb.CreateGuideResponse, error) {
	if req.Content.Before != InitCommitID {
		return &pb.CreateGuideResponse{}, nil
	}

	orgID, err := strconv.ParseUint(req.OrgID, 10, 64)
	if err != nil {
		return nil, apierrors.ErrCreateGuide.InvalidParameter(err)
	}
	projectID, err := strconv.ParseUint(req.ProjectID, 10, 64)
	if err != nil {
		return nil, apierrors.ErrCreateGuide.InvalidParameter(err)
	}
	appID, err := strconv.ParseUint(req.ApplicationID, 10, 64)
	if err != nil {
		return nil, apierrors.ErrCreateGuide.InvalidParameter(err)
	}
	guide := db.Guide{
		// TODO add jump link
		JumpLink:      "",
		Status:        db.InitStatus.String(),
		Kind:          db.PipelineGuide.String(),
		Creator:       req.Content.Pusher.ID,
		OrgID:         orgID,
		ProjectID:     projectID,
		AppID:         appID,
		Branch:        req.Content.Ref[len(BranchPrefix):],
		SoftDeletedAt: 0,
	}
	if err = g.db.CreateGuide(&guide); err != nil {
		return nil, apierrors.ErrCreateGuide.InternalError(err)
	}
	return &pb.CreateGuideResponse{Data: guide.Convert()}, nil
}

func (g *GuideService) ListGuide(ctx context.Context, req *pb.ListGuideRequest) (*pb.ListGuideResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, apierrors.ErrListGuide.InvalidParameter(err)
	}
	guidesDB, err := g.db.ListGuide(req, apis.GetUserID(ctx))
	if err != nil {
		return nil, apierrors.ErrListGuide.InternalError(err)
	}
	guides := make([]*pb.Guide, 0, len(guidesDB))
	for _, v := range guidesDB {
		guides = append(guides, v.Convert())
	}
	return &pb.ListGuideResponse{Data: guides}, nil
}

func (g *GuideService) JudgeCanCreatePipeline(ctx context.Context, req *pb.JudgeCanCreatePipelineRequest) (*pb.JudgeCanCreatePipelineResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, apierrors.ErrJudgeCanCreatePipeline.InvalidParameter(err)
	}

	guideDB, err := g.db.GetGuide(req.ID)
	if err != nil {
		return nil, apierrors.ErrJudgeCanCreatePipeline.InternalError(err)
	}

	pipelineYmls, err := g.ListPipelineYml(ctx, guideDB.AppID, guideDB.Branch)
	if err != nil {
		return nil, apierrors.ErrJudgeCanCreatePipeline.InternalError(err)
	}

	var (
		path      string
		fileName  string
		canCreate bool
	)

	// pipeline.yml first
	sort.Slice(pipelineYmls, func(i, j int) bool {
		return pipelineYmls[i].YmlPath < pipelineYmls[j].YmlPath
	})
	if len(pipelineYmls) > 0 {
		canCreate = true
		path = pipelineYmls[0].YmlPath
		fileName = pipelineYmls[0].YmlName
	}

	return &pb.JudgeCanCreatePipelineResponse{
		CanCreate: canCreate,
		AppID:     guideDB.AppID,
		Branch:    guideDB.Branch,
		Path:      path,
		FileName:  fileName,
	}, nil
}

func (g *GuideService) ProcessGuide(ctx context.Context, req *pb.ProcessGuideRequest) (*pb.ProcessGuideResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, apierrors.ErrProcessGuide.InvalidParameter(err)
	}

	if req.Kind == db.PipelineGuide.String() {
		if err := g.db.UpdateGuideByAppIDAndBranch(req.AppID, req.Branch, req.Kind, map[string]interface{}{"status": db.ProcessedStatus}); err != nil {
			return nil, apierrors.ErrProcessGuide.InternalError(err)
		}
	}

	return &pb.ProcessGuideResponse{}, nil
}

func (g *GuideService) BatchUpdateGuideExpiryStatus() error {
	return g.db.BatchUpdateGuideExpiryStatus()
}

type PipelineYml struct {
	YmlName string
	YmlPath string
}

func (g *GuideService) ListPipelineYml(ctx context.Context, appID uint64, branch string) ([]PipelineYml, error) {
	app, err := g.bdl.GetApp(appID)
	if err != nil {
		return nil, err
	}

	work := limit_sync_group.NewWorker(3)
	var list []PipelineYml
	var pathList = []string{"", DicePipelinePath, ErdaPipelinePath}
	for _, path := range pathList {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			result, err := g.getPipelineYml(app, apis.GetUserID(ctx), branch, i[0].(string))
			if err != nil {
				return nil
			}

			locker.Lock()
			defer locker.Unlock()
			list = append(list, result...)
			return nil
		}, path)
	}
	if err := work.Do().Error(); err != nil {
		return nil, err
	}

	return list, nil
}

func (g *GuideService) getPipelineYml(app *apistructs.ApplicationDTO, userID string, branch string, findPath string) ([]PipelineYml, error) {
	var path string
	if findPath == "" {
		path = fmt.Sprintf("/wb/%v/%v/tree/%v", app.ProjectName, app.Name, branch)
	} else {
		path = fmt.Sprintf("/wb/%v/%v/tree/%v/%v", app.ProjectName, app.Name, branch, findPath)
	}

	diceEntries, err := g.bdl.GetGittarTreeNode(path, strconv.Itoa(int(app.OrgID)), true, userID)
	if err != nil {
		return nil, err
	}

	var list []PipelineYml
	for _, entry := range diceEntries {
		if !strings.HasSuffix(entry.Name, ".yml") {
			continue
		}
		if findPath == "" && entry.Name != apistructs.DefaultPipelineYmlName {
			continue
		}
		list = append(list, PipelineYml{
			YmlName: entry.Name,
			YmlPath: findPath,
		})
	}
	return list, nil
}
