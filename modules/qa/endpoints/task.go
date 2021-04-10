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

package endpoints

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/pipeline/pipelineyml"
	"github.com/erda-project/erda/pkg/httpclient"
)

const (
	BranchPrefix = "refs/heads/"
)

var TaskQueue = make(chan *apistructs.GittarPushEventRequest, 300)

func StartHookTaskConsumer() {
	var (
		err      error
		pipeline *pipelineyml.Pipeline
		content  []byte
	)

	for {
		req := <-TaskQueue

		branchName := strings.TrimPrefix(req.Ref, BranchPrefix)
		if req.Repository == nil {
			logrus.Errorf("nil point repository")
			continue
		}

		if branchName != "develop" {
			logrus.Infof("no need to do sonar and ut, branch: %s", branchName)
			continue
		}

		if req.After == "" {
			logrus.Errorf("nil commit_id")
			continue
		}

		yml := &pipelineyml.PipelineYml{}
		if pipeline, err = yml.CreatePipeline(req.Repository.URL, branchName, req.After); err != nil {
			logrus.Errorf("failed to create pipeline, (%+v)", err)
			continue
		}

		if content, err = yaml.Marshal(pipeline); err != nil {
			logrus.Errorf("failed to marshal pipeline, (%+v)", err)
			continue
		}

		_, err = CreateBuild(string(content), branchName, req.Pusher.ID, uint64(req.Repository.ApplicationID))
		if err != nil {
			logrus.Errorf("failed to create pipeline, (%+v)", err)
		}
	}
}

func CreateBuild(pipeline, branch, uid string, appId uint64) (*apistructs.QaBuildCreateResponse, error) {
	req := apistructs.PipelineCreateRequest{
		AppID:              appId,
		Branch:             branch,
		Source:             "qa",
		PipelineYmlSource:  "content",
		PipelineYmlName:    "qa.yml",
		PipelineYmlContent: pipeline,
		AutoRun:            true,
	}

	openApiToken, err := getOpenapiToken()
	if err != nil {
		return nil, err
	}

	var result apistructs.QaBuildCreateResponse
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Post(os.Getenv("OPENAPI_PUBLIC_URL")).
		Path("/api/pipelines").
		Header("Content-Type", "application/json").
		Header("Authorization", openApiToken).
		Header("User-ID", uid).
		JSONBody(&req).
		Do().
		JSON(&result)

	if err != nil {
		return nil, errors.Errorf("failed to create build, req: %+v", req)
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to create build, code: %d, req: %+v",
			r.StatusCode(), req)
	}

	if !result.Success {
		return nil, errors.Errorf("failed to create build, code: %s, msg:%s",
			result.Error.Code, result.Error.Msg)
	}

	return &result, nil
}

func getOpenapiToken() (string, error) {
	var result struct {
		AccessToken string `json:"access_token"`
	}
	authHeader := base64.StdEncoding.EncodeToString([]byte("pipeline:devops/pipeline"))
	resp, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Post(os.Getenv("OPENAPI_PUBLIC_URL")).
		Path("/api/openapi/client-token").
		Header("Authorization", "Basic "+authHeader).
		Do().
		JSON(&result)
	if err != nil {
		return "", errors.Errorf("failed to get openapi, (%+v)", err)
	}
	if !resp.IsOK() {
		return "", errors.Errorf("failed to get openapi, statusCode: %d", resp.StatusCode())
	}

	return result.AccessToken, nil
}
