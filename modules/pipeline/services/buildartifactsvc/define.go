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

package buildartifactsvc

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

type BuildArtifactSvc struct {
	dbClient *dbclient.Client
}

func New(dbClient *dbclient.Client) *BuildArtifactSvc {
	s := BuildArtifactSvc{}
	s.dbClient = dbClient
	return &s
}

func (s *BuildArtifactSvc) Query(sha256 string) (*spec.CIV3BuildArtifact, error) {
	artifact, err := s.dbClient.GetBuildArtifactBySha256(sha256)
	if err != nil {
		return nil, apierrors.ErrQueryBuildArtifact.InternalError(err)
	}
	if artifact.Type == apistructs.BuildArtifactOfFileContent {
		// 解析 artifact 内容，获取镜像列表，查询 dicehub
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
				if r.StatusCode() == http.StatusNotFound { // dicehub 明确告知镜像已经不存在，则删除这条 artifact 记录
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
	return &artifact, nil
}

func (s *BuildArtifactSvc) Register(req *apistructs.BuildArtifactRegisterRequest) (*spec.CIV3BuildArtifact, error) {
	artifact, err := s.dbClient.NewArtifact(
		req.SHA, req.IdentityText, apistructs.BuildArtifactType(req.Type),
		req.Content, req.ClusterName, req.PipelineID,
	)
	if err != nil {
		return nil, apierrors.ErrRegisterBuildArtifact.InternalError(err)
	}
	return &artifact, nil
}

func (s *BuildArtifactSvc) Delete(req *apistructs.BuildArtifactDeleteByImagesRequest) error {
	if req.Images == nil || len(req.Images) == 0 {
		return nil
	}

	for _, image := range req.Images {
		if !strutil.Contains(image, ":") || !strutil.Contains(image, "/") || len(image) < 15 {
			return apierrors.ErrDeleteBuildArtifact.InvalidParameter(strutil.Concat("invalid image: ", image))
		}
	}

	if err := s.dbClient.DeleteArtifactsByImages(apistructs.BuildArtifactOfFileContent, req.Images); err != nil {
		return apierrors.ErrDeleteBuildArtifact.InternalError(err)
	}
	return nil
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
