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

// Package application 应用逻辑封装
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/core-services/types"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

// Application 应用操作封装
type Application struct {
	db  *dao.DBClient
	uc  *ucauth.UCClient
	bdl *bundle.Bundle
	cms cmspb.CmsServiceServer
}

// Option 定义 Appliction 对象的配置选项
type Option func(*Application)

// New 新建 Application 实例，操作应用资源
func New(options ...Option) *Application {
	app := &Application{}
	for _, op := range options {
		op(app)
	}
	return app
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(a *Application) {
		a.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(a *Application) {
		a.uc = uc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Application) {
		a.bdl = bdl
	}
}

func WithPipelineCms(cms cmspb.CmsServiceServer) Option {
	return func(a *Application) {
		a.cms = cms
	}
}

// CreateWithEvent 创建应用 & 发送事件
func (a *Application) CreateWithEvent(userID string, createReq *apistructs.ApplicationCreateRequest) (*model.Application, error) {
	// 创建应用
	if createReq.DisplayName == "" {
		createReq.DisplayName = createReq.Name
	}
	app, err := a.Create(userID, createReq)
	if err != nil {
		return app, err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ApplicationEvent,
			Action:        bundle.CreateAction,
			OrgID:         strconv.FormatInt(app.OrgID, 10),
			ProjectID:     strconv.FormatInt(app.ProjectID, 10),
			ApplicationID: strconv.FormatInt(app.ID, 10),
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *app,
	}
	// 发送应用创建事件
	if err = a.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send application create event, (%v)", err)
	}

	return app, nil
}

func getRepoConfigStr(config *apistructs.GitRepoConfig) string {
	configBytes, _ := json.Marshal(map[string]string{
		"url":  config.Url,
		"desc": config.Desc,
		"type": config.Type,
	})
	return string(configBytes)
}

// Create 创建应用
func (a *Application) Create(userID string, createReq *apistructs.ApplicationCreateRequest) (*model.Application, error) {
	// 检查name重名
	app, err := a.db.GetApplicationByName(int64(createReq.ProjectID), createReq.Name)
	if err != nil {
		logrus.Warnf("failed to get application, (%v)", err)
		return nil, errors.Errorf("failed to get application")
	}
	if app.ID > 0 {
		return nil, errors.Errorf("failed to create application(name already exists)")
	}

	config, err := json.Marshal(createReq.Config)
	if err != nil {
		return nil, errors.Errorf("failed to marshal config, (%v)", err)
	}

	// 生成gitRepoAbbrev & gitRepo
	project, err := a.db.GetProjectByID(int64(createReq.ProjectID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create application")
	}
	org, _ := a.db.GetOrg(project.OrgID)

	// 添加application至DB
	application := model.Application{
		Name:           createReq.Name,
		DisplayName:    createReq.DisplayName,
		Desc:           createReq.Desc,
		Logo:           createReq.Logo,
		Config:         string(config),
		OrgID:          org.ID,
		ProjectID:      int64(createReq.ProjectID),
		ProjectName:    project.Name,
		Mode:           string(createReq.Mode),
		UserID:         userID,
		IsExternalRepo: createReq.IsExternalRepo,
	}
	// 关联gittar
	repoReq := apistructs.CreateRepoRequest{
		OrgID:       org.ID,
		OrgName:     org.Name,
		ProjectID:   project.ID,
		ProjectName: project.Name,
		AppName:     application.Name,
		IsExternal:  createReq.IsExternalRepo,
		Config:      createReq.RepoConfig,
	}
	// 外置仓库先尝试检测一下有效性
	if createReq.IsExternalRepo {
		repoReq.OnlyCheck = true
		_, err := a.bdl.CreateRepo(repoReq)
		if err != nil {
			return nil, errors.Errorf("failed to create repo, (%v)", err)
		}
		application.RepoConfig = getRepoConfigStr(createReq.RepoConfig)
	}
	if err = a.db.CreateApplication(&application); err != nil {
		logrus.Warnf("failed to insert application to db, (%v)", err)
		return nil, errors.Errorf("failed to insert application to db")
	}

	repoReq.OnlyCheck = false
	repoReq.AppID = application.ID
	gittarResp, err := a.bdl.CreateRepo(repoReq)
	if err != nil {
		a.db.DeleteApplication(application.ID)
		a.bdl.DeleteRepo(application.ID)
		return nil, errors.Errorf("failed to create repo, (%v)", err)
	}

	// 更新extra等信息
	application.Extra = a.generateExtraInfo(application.ID, application.ProjectID)
	application.GitRepoAbbrev = gittarResp.RepoPath
	if err = a.db.UpdateApplication(&application); err != nil {
		logrus.Errorf("failed to update application extra, (%v)", err)
	}

	// 新增应用管理员至admin_members表
	users, err := a.uc.FindUsers([]string{userID})
	if err != nil {
		logrus.Warnf("failed to get user info, (%v)", err)
	}
	if len(users) > 0 {
		member := model.Member{
			ScopeType:     apistructs.AppScope,
			ScopeID:       application.ID,
			ScopeName:     application.Name,
			ParentID:      application.ProjectID,
			UserID:        userID,
			Email:         users[0].Email,
			Mobile:        users[0].Phone,
			Name:          users[0].Name,
			Nick:          users[0].Nick,
			Avatar:        users[0].AvatarURL,
			UserSyncAt:    time.Now(),
			OrgID:         org.ID,
			ProjectID:     application.ProjectID,
			ApplicationID: application.ID,
			Token:         uuid.UUID(),
		}
		memberExtra := model.MemberExtra{
			UserID:        userID,
			ScopeID:       application.ID,
			ScopeType:     apistructs.AppScope,
			ParentID:      application.ProjectID,
			ResourceKey:   apistructs.RoleResourceKey,
			ResourceValue: types.RoleAppOwner,
		}
		if err = a.db.CreateMember(&member); err != nil {
			logrus.Warnf("failed to add member, (%v)", err)
		}
		if err = a.db.CreateMemberExtra(&memberExtra); err != nil {
			logrus.Warnf("failed to add member roles to db when create project, (%v)", err)
		}
	}

	return &application, nil
}

// 创建应用时，自动生成extra信息
func (a *Application) generateExtraInfo(applicationID, projectID int64) string {
	// 初始化DEV、TEST、STAGING、PROD四个环境namespace，eg: "DEV.configNamespace":"app-107-DEV"
	workspaces := []apistructs.DiceWorkspace{
		types.DefaultWorkspace,
		types.DevWorkspace,
		types.TestWorkspace,
		types.StagingWorkspace,
		types.ProdWorkspace,
	}

	extra := make(map[string]string, len(workspaces))
	for _, v := range workspaces {
		key := strutil.Concat(string(v), ".configNamespace")
		value := strutil.Concat("app-", strconv.FormatInt(applicationID, 10), "-", string(v))
		extra[key] = value
	}

	extraInfo, err := json.Marshal(extra)
	if err != nil {
		logrus.Errorf("failed to marshal extra info, (%v)", err)
	}
	return string(extraInfo)
}

// UpdateWithEvent 更新应用 & 发送事件
func (a *Application) UpdateWithEvent(appID int64, updateReq *apistructs.ApplicationUpdateRequestBody) (*model.Application, error) {
	// 更新应用
	app, err := a.Update(appID, updateReq)
	if err != nil {
		return nil, err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ApplicationEvent,
			Action:        bundle.UpdateAction,
			OrgID:         strconv.FormatInt(app.OrgID, 10),
			ProjectID:     strconv.FormatInt(app.ProjectID, 10),
			ApplicationID: strconv.FormatInt(app.ID, 10),
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *app,
	}
	// 发送应用更新事件
	if err = a.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send application update event, (%v)", err)
	}

	return app, nil
}

// Init 应用初始化
func (a *Application) Init(initReq *apistructs.ApplicationInitRequest) (uint64, error) {
	// 1. 参数校验
	// 2. 生成 pipeline.yml
	// 3. 调用 pipeline API 创建 pipeline
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

	app, err := a.db.GetApplicationByID(int64(initReq.ApplicationID))
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
	if members, err := a.db.GetMemberByScopeAndUserID(initReq.UserID, apistructs.OrgScope, app.OrgID); err == nil && members != nil {
		// TODO: kanbudong
		token = members[0].Token
	}
	if token == "" {
		return 0, errors.Errorf("not found user token")
	}
	org, err := a.db.GetOrg(app.OrgID)
	if err != nil {
		return 0, err
	}
	domain := strutil.Concat(strutil.ToLower(org.Name), "-org.", conf.RootDomain())
	u, _ := url.Parse(conf.UIPublicURL())
	remoteUrl := fmt.Sprintf("%s://git:%s@%s/wb/%s/%s", u.Scheme, token, domain, app.ProjectName, app.Name)

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
	project, err := a.db.GetProjectByID(app.ProjectID)
	if err != nil {
		return 0, err
	}
	clusterConfig := map[string]string{}
	json.Unmarshal([]byte(project.ClusterConfig), &clusterConfig)
	clusterName, ok := clusterConfig[string(apistructs.DevWorkspace)]
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
			apistructs.LabelOrgID:         strconv.FormatInt(app.OrgID, 10),
			apistructs.LabelProjectID:     strconv.FormatInt(app.ProjectID, 10),
			apistructs.LabelAppID:         strconv.FormatInt(app.ID, 10),
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

// Update 更新应用
func (a *Application) Update(appID int64, updateReq *apistructs.ApplicationUpdateRequestBody) (
	*model.Application, error) {
	// 检查应用是否存在
	application, err := a.db.GetApplicationByID(appID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update application")
	}
	if updateReq.DisplayName == "" {
		updateReq.DisplayName = application.Name
	}
	// 更新应用信息
	application.Desc = updateReq.Desc
	application.Logo = updateReq.Logo
	application.DisplayName = updateReq.DisplayName
	application.IsPublic = updateReq.IsPublic
	config, err := json.Marshal(updateReq.Config)
	if err != nil {
		return nil, errors.Errorf("failed to marshal config, (%v)", err)
	}
	application.Config = string(config)
	if application.IsExternalRepo && updateReq.RepoConfig != nil {
		application.RepoConfig = getRepoConfigStr(updateReq.RepoConfig)
		err = a.bdl.UpdateRepo(apistructs.UpdateRepoRequest{
			AppID:  application.ID,
			Config: updateReq.RepoConfig,
		})
		if err != nil {
			return nil, errors.Errorf("failed to update repo, (%v)", err)
		}
	}
	if err = a.db.UpdateApplication(&application); err != nil {
		logrus.Warnf("failed to update application, (%v)", err)
		return nil, errors.Errorf("failed to update application")
	}

	return &application, nil
}

// Pin pin 应用
func (a *Application) Pin(appID int64, userID string) error {
	// 检查应用是否存在
	app, err := a.db.GetApplicationByID(appID)
	if err != nil {
		return err
	}

	// 若已收藏，则返回
	fr, err := a.db.GetFavoritedResource(string(apistructs.AppScope), uint64(appID), userID)
	if err != nil {
		return err
	}
	if fr != nil {
		return nil
	}

	// 添加收藏关系
	fr = &model.FavoritedResource{
		Target:   string(apistructs.AppScope),
		TargetID: uint64(appID),
		UserID:   userID,
	}
	if err := a.db.CreateFavoritedResource(fr); err != nil {
		return errors.Wrap(err, "failed to pin application")
	}
	app.UpdatedAt = time.Now()
	if err := a.db.UpdateApplication(&app); err != nil {
		return errors.Wrap(err, "failed to pin application")
	}

	return nil
}

// UnPin unpin 应用
func (a *Application) UnPin(appID int64, userID string) error {
	// 检查应用是否存在
	_, err := a.db.GetApplicationByID(appID)
	if err != nil {
		return err
	}

	// 若未收藏，则返回
	fr, err := a.db.GetFavoritedResource(string(apistructs.AppScope), uint64(appID), userID)
	if err != nil {
		return err
	}
	if fr == nil {
		return nil
	}

	// 删除收藏关系
	if err := a.db.DeleteFavoritedResource(uint64(fr.ID)); err != nil {
		return errors.Wrap(err, "failed to unpin application")
	}

	return nil
}

// DeleteWithEvent 删除应用 & 发送事件
func (a *Application) DeleteWithEvent(applicationID int64) error {
	// 删除应用
	app, err := a.Delete(applicationID)
	if err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ApplicationEvent,
			Action:        bundle.DeleteAction,
			OrgID:         strconv.FormatInt(app.OrgID, 10),
			ProjectID:     strconv.FormatInt(app.ProjectID, 10),
			ApplicationID: strconv.FormatInt(app.ID, 10),
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *app,
	}
	// 发送应用删除事件
	if err = a.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send application update event, (%v)", err)
	}

	return nil
}

// Delete 删除应用
func (a *Application) Delete(applicationID int64) (*model.Application, error) {
	// 查询应用是否存在
	application, err := a.db.GetApplicationByID(applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete application")
	}

	// 判断应用下是否有runtime，无 runtime 时应用可删除
	var runtimeCount int64
	a.db.Table("ps_v2_project_runtimes").Where("application_id = ?", applicationID).Count(&runtimeCount)
	if runtimeCount > 0 {
		return nil, errors.Errorf("failed to delete application(there exists runtime)")
	}

	// 删除gittar repo
	if err = a.bdl.DeleteRepo(application.ID); err != nil {
		// 防止有老数据不存在repoID，还是以repo路径进行删除
		if err = a.bdl.DeleteGitRepo(application.GitRepoAbbrev); err != nil {
			logrus.Errorf(err.Error())
		}
	}

	// 删除应用
	if err = a.db.DeleteApplication(applicationID); err != nil {
		logrus.Warnf("failed to delete application, (%v)", err)
		return nil, errors.Errorf("failed to delete application")
	}
	logrus.Infof("deleted application: %d", applicationID)

	// 删除应用下成员及权限
	if err := a.db.DeleteMembersByScope(apistructs.AppScope, applicationID); err != nil {
		logrus.Warnf("failed to delete members, (%v)", err)
	}
	if err := a.db.DeleteMemberExtraByScope(apistructs.AppScope, applicationID); err != nil {
		logrus.Warnf("failed to delete members extra, (%v)", err)
	}

	// 删除此应用所有的收藏关系
	if err := a.db.DeleteFavoritedResourcesByTarget(string(apistructs.AppScope), uint64(applicationID)); err != nil {
		logrus.Warnf("failed to delete app FavoritedResources, (%v)", err)
	}

	if err = a.db.DeleteNotifyRelationsByScope(apistructs.AppScope, strconv.FormatInt(applicationID, 10)); err != nil {
		logrus.Warnf("failed to delete notify relations, (%v)", err)
	}

	return &application, nil
}

// Get 获取应用
func (a *Application) Get(applicationID int64) (*model.Application, error) {
	application, err := a.db.GetApplicationByID(applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get application")
	}
	if application.DisplayName == "" {
		application.DisplayName = application.Name
	}
	return &application, nil
}

// GetAllAppsByProject 根据projectID 获取应用
func (a *Application) GetAllAppsByProject(projectID int64) ([]model.Application, error) {
	return a.db.GetProjectApplications(projectID)
}

// GetAllApps 获取所有app列表
func (a *Application) GetAllApps() ([]model.Application, error) {
	return a.db.GetAllApps()
}

// List 应用列表/查询
func (a *Application) List(orgID, projectID int64, userID string, request *apistructs.ApplicationListRequest) (
	int, []model.Application, error) {
	// 获取应用列表
	applicationIDs := make([]int64, 0)
	total, applications, err := a.db.GetApplicationsByIDs(&orgID, &projectID, applicationIDs, request)
	if err != nil {
		logrus.Infof("failed to get application list, (%v)", err)
		return 0, nil, errors.Errorf("failed to get application list")
	}

	frs, err := a.db.GetFavoritedResourcesByUser(string(apistructs.AppScope), userID)
	if err != nil {
		return 0, nil, err
	}
	frMap := make(map[uint64]model.FavoritedResource, len(frs))
	for _, v := range frs {
		frMap[v.TargetID] = v
	}

	// 若存在收藏关系，则讲 pined 置为 true
	for i := range applications {
		if applications[i].DisplayName == "" {
			applications[i].DisplayName = applications[i].Name
		}
		if _, ok := frMap[uint64(applications[i].ID)]; ok {
			applications[i].Pined = true
		}
	}

	return total, applications, nil
}

// ListMyApplications 我的应用列表
func (a *Application) ListMyApplications(orgID int64, userID string, request *apistructs.ApplicationListRequest) (
	int, []model.Application, error) {
	var (
		members []model.MemberExtra
		err     error
	)

	// 查找有权限的列表
	members, err = a.db.GetAppMembersByUser(userID)
	if err != nil {
		logrus.Infof("failed to get permission, (%v)", err)
		return 0, nil, errors.Errorf("failed to get permission")
	}

	// 获取应用列表
	applicationIDs := make([]int64, 0, len(members)*2)
	for i := range members {
		applicationIDs = append(applicationIDs, members[i].ScopeID)
	}

	// 用户没加入任何项目和应用
	if len(applicationIDs) == 0 {
		return 0, nil, nil
	}

	total, applications, err := a.db.GetApplicationsByIDs(&orgID, nil, applicationIDs, request)
	if err != nil {
		logrus.Infof("failed to get application list, (%v)", err)
		return 0, nil, errors.Errorf("failed to get application list")
	}

	frs, err := a.db.GetFavoritedResourcesByUser(string(apistructs.AppScope), userID)
	if err != nil {
		return 0, nil, err
	}
	frMap := make(map[uint64]model.FavoritedResource, len(frs))
	for _, v := range frs {
		frMap[v.TargetID] = v
	}

	// 若存在收藏关系，则讲 pined 置为 true
	for i := range applications {
		if applications[i].DisplayName == "" {
			applications[i].DisplayName = applications[i].Name
		}
		if _, ok := frMap[uint64(applications[i].ID)]; ok {
			applications[i].Pined = true
		}
	}

	return total, applications, nil
}

// ListByProjectID 根据projectID获取应用列表
func (a *Application) ListByProjectID(projectID, pageNum, pageSize int64) ([]model.Application, error) {
	applications, err := a.db.GetApplicationsByProjectID(projectID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range applications {
		if applications[i].DisplayName == "" {
			applications[i].DisplayName = applications[i].Name
		}
	}
	return applications, nil
}

// QueryPublishItemRelations 查询应用发布内容关联关系
func (a *Application) QueryPublishItemRelations(req apistructs.QueryAppPublishItemRelationRequest) ([]apistructs.AppPublishItemRelation, error) {
	relations, err := a.db.QueryAppPublishItemRelations(req)
	if err != nil {
		return nil, err
	}

	return relations, nil
}

// UpdatePublishItemRelations 增量更新或创建publishItemRelations
func (a *Application) UpdatePublishItemRelations(request *apistructs.UpdateAppPublishItemRelationRequest) error {
	relations, err := a.db.QueryAppPublishItemRelations(apistructs.QueryAppPublishItemRelationRequest{AppID: request.AppID})
	if err != nil {
		return err
	}
	result := map[string]apistructs.AppPublishItemRelation{}
	for _, relation := range relations {
		var itemNs []string
		itemNs = append(itemNs, a.BuildItemMonitorPipelineCmsNs(relation.AppID, relation.Env))
		relation.PublishItemNs = itemNs
		result[relation.Env] = relation
	}

	app, err := a.db.GetApplicationByID(request.AppID)
	if err != nil {
		return err
	}

	for _, workspace := range apistructs.DiceWorkspaceSlice {
		// 获取AK(TK)
		monitorAddon, err := a.bdl.ListByAddonName("monitor", strconv.FormatInt(app.ProjectID, 10), workspace.String())
		if err != nil {
			return err
		}
		if len(monitorAddon.Data) == 0 {
			return errors.Errorf("the monitor addon doesn't exist ENV: %s, projectID: %d", workspace.String(), app.ProjectID)
		}
		AK, ok := monitorAddon.Data[0].Config["TERMINUS_KEY"]
		if !ok {
			return errors.Errorf("the monitor addon doesn't have TERMINUS_KEY")
		}
		request.AKAIMap[workspace] = apistructs.MonitorKeys{AK: AK.(string), AI: app.Name}
		// 更新app relation
		relation, ok := result[workspace.String()]
		if ok && request.GetPublishItemIDByWorkspace(workspace) == relation.PublishItemID &&
			request.AppID == relation.AppID && relation.AK == AK.(string) && relation.AI == app.Name {
			// 数据库已经存在记录，且不需要更新
			request.SetPublishItemIDTo0ByWorkspace(workspace)
		}
	}

	if err := a.db.UpdateAppPublishItemRelations(request); err != nil {
		return err
	}

	return a.PipelineCmsConfigRequest(request)
}

// PipelineCmsConfigRequest 请求pipeline cms，将publisherKey和publishItemKey设置进配置管理
func (a *Application) PipelineCmsConfigRequest(request *apistructs.UpdateAppPublishItemRelationRequest) error {
	for workspace, mk := range request.AKAIMap {
		// bundle req
		if _, err := a.cms.UpdateCmsNsConfigs(apis.WithInternalClientContext(context.Background(), "core-services"),
			&cmspb.CmsNsConfigsUpdateRequest{
				Ns:             a.BuildItemMonitorPipelineCmsNs(request.AppID, workspace.String()),
				PipelineSource: apistructs.PipelineSourceDice.String(),
				KVs:            map[string]*cmspb.PipelineCmsConfigValue{"AI": {Value: mk.AI}, "AK": {Value: mk.AK}},
			}); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) RemovePublishItemRelations(request *apistructs.RemoveAppPublishItemRelationsRequest) error {
	return a.db.RemovePublishItemRelations(request)
}

// BuildItemMonitorPipelineCmsNs 生成namespace
func (a *Application) BuildItemMonitorPipelineCmsNs(appID int64, workspace string) string {
	return fmt.Sprintf("publish-item-monitor-%s-%d", workspace, appID)
}
