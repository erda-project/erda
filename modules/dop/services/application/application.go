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

package application

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

type Application struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

type Option func(*Application)

func New(options ...Option) *Application {
	app := &Application{}
	for _, op := range options {
		op(app)
	}
	return app
}

func WithDBClient(db *dao.DBClient) Option {
	return func(a *Application) {
		a.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Application) {
		a.bdl = bdl
	}
}

func (a *Application) Init(initReq *apistructs.ApplicationInitRequest) (uint64, error) {
	if initReq.MobileAppName == "" {
		return 0, apierrors.ErrInitApplication.MissingParameter("mobileAppName")
	}
	if initReq.BundleID == "" {
		return 0, apierrors.ErrInitApplication.MissingParameter("bundleID")
	}
	if initReq.PackageName == "" {
		return 0, apierrors.ErrInitApplication.MissingParameter("packageName")
	}
	if initReq.MobileDisplayName == "" {
		initReq.MobileDisplayName = initReq.MobileAppName
	}

	app, err := a.bdl.GetApp(initReq.ApplicationID)
	if err != nil {
		return 0, err
	}
	if app.Mode != string(apistructs.ApplicationModeMobile) {
		return 0, errors.Errorf("only support mobile app template init")
	}

	actionProjectName := strings.Replace(initReq.MobileAppName, "-", "", -1)
	// generate mobile template action
	mobileTemplateStage := &apistructs.PipelineYmlAction{
		Type: "mobile-template",
		Params: map[string]interface{}{
			"project_name": actionProjectName,
			"display_name": initReq.MobileDisplayName,
			"bundle_id":    initReq.BundleID,
			"package_name": initReq.PackageName,
		},
	}

	// generate remote url
	var token string
	if members, err := a.bdl.GetMemberByUserAndScope(apistructs.OrgScope, initReq.UserID, app.OrgID); err == nil && members != nil {
		token = members[0].Token
	}
	if token == "" {
		return 0, errors.Errorf("not found user token")
	}
	org, err := a.bdl.GetOrg(app.OrgID)
	if err != nil {
		return 0, err
	}
	u, _ := url.Parse(conf.UIPublicURL())
	remoteUrl := fmt.Sprintf("%s://git:%s@%s/%s/dop/%s/%s", u.Scheme, token, conf.UIDomain(), org.Name, app.ProjectName, app.Name)

	// generate git push action
	gitPushStage := &apistructs.PipelineYmlAction{
		Type: "git-push",
		Params: map[string]interface{}{
			"workdir":    "${mobile-template}/" + actionProjectName,
			"remote_url": remoteUrl,
		},
	}

	// generate pipeline.yml
	pys := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{
			{
				mobileTemplateStage,
			},
			{
				gitPushStage,
			},
		},
	}

	// fetch cluster name
	project, err := a.bdl.GetProject(app.ProjectID)
	if err != nil {
		return 0, err
	}
	clusterName, ok := project.ClusterConfig[string(apistructs.DevWorkspace)]
	if !ok {
		return 0, errors.Errorf("not found cluster")
	}

	ymlContent, err := yaml.Marshal(&pys)
	if err != nil {
		return 0, err
	}
	req := &apistructs.PipelineCreateRequestV2{
		PipelineYml:     string(ymlContent),
		PipelineYmlName: fmt.Sprintf("%s_%s_pipeline.yml", project.Name, app.Name),
		PipelineSource:  apistructs.PipelineSourceDice,
		ClusterName:     clusterName,
		Labels: map[string]string{
			apistructs.LabelBranch:        "master",
			apistructs.LabelOrgID:         strconv.FormatUint(app.OrgID, 10),
			apistructs.LabelProjectID:     strconv.FormatUint(app.ProjectID, 10),
			apistructs.LabelAppID:         strconv.FormatUint(app.ID, 10),
			apistructs.LabelDiceWorkspace: string(apistructs.ProdWorkspace),
		},
		AutoRunAtOnce: true,
	}
	// create pipeline info
	pipelineInfo, err := a.bdl.CreatePipeline(req)
	if err != nil {
		return 0, err
	}

	return pipelineInfo.ID, nil
}
