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

package build

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/pipeline/build/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/build/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

type buildService struct {
	p        *provider
	dbClient *db.Client
}

func (s *buildService) QueryBuildArtifact(ctx context.Context, req *pb.BuildArtifactQueryRequest) (*pb.BuildArtifactQueryResponse, error) {
	artifact, err := s.dbClient.GetBuildArtifactBySha256(req.Sha)
	if err != nil {
		return nil, apierrors.ErrQueryBuildArtifact.InternalError(err)
	}
	if artifact.Type == apistructs.BuildArtifactOfFileContent {
		// Parse artifact content, get mirror list, query dicehub
		images, err := parseImagesFromContent(artifact.Content)
		if err != nil {
			return nil, apierrors.ErrQueryBuildArtifact.InternalError(err)
		}
		for _, image := range images {
			var body apistructs.ImageGetResponse
			r, err := httpclient.New().Get(discover.DiceHub()).Path("/api/images/" + url.PathEscape(image)).Do().JSON(&body)
			if err != nil {
				return nil, apierrors.ErrQueryDicehub.InternalError(err)
			}
			if !r.IsOK() {
				if r.StatusCode() == http.StatusNotFound { // dicehub clearly informs that the mirror no longer exists, then delete the artifact record
					if err := s.dbClient.DeleteArtifact(artifact.ID); err != nil {
						logrus.Errorf("[alert] failed to delete artifact which image is not found in dicehub, image: %s", image)
					}
				}
				return nil, apierrors.ErrQueryDicehub.InternalError(errors.Errorf("%+v", body.Error))
			}
			if !body.Success {
				return nil, apierrors.ErrQueryBuildArtifact.InvalidState(strutil.Concat("parsed image out of date: ", image))
			}
		}
	}
	return &pb.BuildArtifactQueryResponse{Data: artifact.Convert2PB()}, nil
}

func (s *buildService) RegisterBuildArtifact(ctx context.Context, req *pb.BuildArtifactRegisterRequest) (*pb.BuildArtifactRegisterResponse, error) {
	artifact, err := s.dbClient.NewArtifact(
		req.Sha, req.IdentityText, apistructs.BuildArtifactType(req.Type),
		req.Content, req.ClusterName, req.PipelineID,
	)
	if err != nil {
		return nil, apierrors.ErrRegisterBuildArtifact.InternalError(err)
	}
	return &pb.BuildArtifactRegisterResponse{
		Data: artifact.Convert2PB(),
	}, nil
}

func (s *buildService) DeleteArtifactsByImages(ctx context.Context, req *pb.BuildArtifactDeleteByImagesRequest) (*pb.BuildArtifactDeleteByImagesResponse, error) {
	if req.Images == nil || len(req.Images) == 0 {
		return nil, nil
	}

	for _, image := range req.Images {
		if !strutil.Contains(image, ":") || !strutil.Contains(image, "/") || len(image) < 15 {
			return nil, apierrors.ErrDeleteBuildArtifact.InvalidParameter(strutil.Concat("invalid image: ", image))
		}
	}

	if err := s.dbClient.DeleteArtifactsByImages(apistructs.BuildArtifactOfFileContent, req.Images); err != nil {
		return nil, apierrors.ErrDeleteBuildArtifact.InternalError(err)
	}
	return &pb.BuildArtifactDeleteByImagesResponse{}, nil
}

func (s *buildService) ReportBuildCache(ctx context.Context, req *pb.BuildCacheReportRequest) (*pb.BuildCacheReportResponse, error) {
	cacheImage := &db.CIV3BuildCache{
		Name:        req.Name,
		ClusterName: req.ClusterName,
	}
	session := s.dbClient.Interface.NewSession()
	defer session.Close()

	success, err := session.Get(cacheImage)
	if err != nil {
		return nil, apierrors.ErrReportBuildCache.InternalError(err)
	}
	if req.Action == "push" {
		if !success {
			if _, err = session.Insert(cacheImage); err != nil {
				return nil, apierrors.ErrReportBuildCache.InternalError(err)
			}
		}
	} else if req.Action == "pull" {
		if success {
			cacheImage.LastPullAt = time.Now()
			if _, err = session.ID(cacheImage.ID).Update(cacheImage); err != nil {
				return nil, apierrors.ErrReportBuildCache.InternalError(err)
			}
		}
	}
	return &pb.BuildCacheReportResponse{}, nil
}

func parseImagesFromContent(content string) ([]string, error) {
	var packResult []ModuleImage
	if err := json.Unmarshal([]byte(content), &packResult); err != nil {
		return nil, err
	}
	var images []string
	for _, packLine := range packResult {
		images = append(images, packLine.Image)
	}
	return images, nil
}

type ModuleImage struct {
	ModuleName string `json:"module_name"`
	Image      string `json:"image"`
}
