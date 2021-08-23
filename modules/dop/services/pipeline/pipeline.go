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

// Package pipeline pipeline相关的结构信息
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/branchrule"
	"github.com/erda-project/erda/modules/dop/services/publisher"
	"github.com/erda-project/erda/modules/dop/utils"
	"github.com/erda-project/erda/modules/pipeline/providers/cms"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/modules/pkg/gitflowutil"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DicePipelinesGitFolder = ".dice/pipelines"
)

// Pipeline pipeline 结构体
type Pipeline struct {
	bdl           *bundle.Bundle
	branchRuleSvc *branchrule.BranchRule
	publisherSvc  *publisher.Publisher
	cms           cmspb.CmsServiceServer
}

// Option Pipeline 配置选项
type Option func(*Pipeline)

// New Pipeline service
func New(options ...Option) *Pipeline {
	r := &Pipeline{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(f *Pipeline) {
		f.bdl = bdl
	}
}

func WithBranchRuleSvc(svc *branchrule.BranchRule) Option {
	return func(f *Pipeline) {
		f.branchRuleSvc = svc
	}
}

func WithPublisherSvc(svc *publisher.Publisher) Option {
	return func(f *Pipeline) {
		f.publisherSvc = svc
	}
}

func WithPipelineCms(cms cmspb.CmsServiceServer) Option {
	return func(f *Pipeline) {
		f.cms = cms
	}
}

// 获取应用下的所有.yml文件
func GetPipelineYmlList(req apistructs.CICDPipelineYmlListRequest, bdl *bundle.Bundle) []string {
	result := []string{}
	files, err := bdl.SearchGittarFiles(req.AppID, req.Branch, "pipeline.yml", "", 1)
	if err == nil {
		for _, file := range files {
			result = append(result, file.Name)
		}
	}

	pipelinePath := DicePipelinesGitFolder
	files, err = bdl.SearchGittarFiles(req.AppID, req.Branch, "*.yml", pipelinePath, 3)
	if err == nil {
		for _, file := range files {
			result = append(result, pipelinePath+"/"+file.Name)
		}
	}

	return result
}

// FetchPipelineYml 获取pipeline.yml文件
func (p *Pipeline) FetchPipelineYml(gittarURL, ref, pipelineYmlName string) (string, error) {
	return p.bdl.GetGittarFile(gittarURL, ref, pipelineYmlName, "", "")
}

// CreatePipeline 创建pipeline流程
func (p *Pipeline) CreatePipeline(reqPipeline *apistructs.PipelineCreateRequest) (*apistructs.PipelineDTO, error) {
	resp, err := p.bdl.CreatePipeline(reqPipeline)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreatePipeline 创建pipeline流程
func (p *Pipeline) CreatePipelineV2(reqPipeline *apistructs.PipelineCreateRequestV2) (*apistructs.PipelineDTO, error) {
	resp, err := p.bdl.CreatePipeline(reqPipeline)
	if err != nil {
		return nil, apierrors.ErrCreatePipeline.InternalError(err)
	}

	return resp, nil
}

// GenerateReleaseYml 根据pipeline.yml生成新的release.yml
func (p *Pipeline) GenerateReleaseYml(strPipelineYml, branchName string) (string, error) {
	pipelineYml, err := pipelineyml.New([]byte(strPipelineYml))
	if err != nil {
		return "", err
	}

	// 解析pipeline，并删除release后面的所有stage，若release action和dice action在同个stage，删除dice action
	var isExistRelease bool
Exit:
	for i, stage := range pipelineYml.Spec().Stages {
		for j, typedAction := range stage.Actions {
			for actionType := range typedAction {
				if actionType.String() == "release" {
					pipelineYml.Spec().Stages = pipelineYml.Spec().Stages[:i+1]
					isExistRelease = true
					break
				}

				// 若release action和dice action在同个stage，删除dice action
				if actionType.String() == "dice" {
					if len(pipelineYml.Spec().Stages[i].Actions) > j+1 {
						pipelineYml.Spec().Stages[i].Actions = append(pipelineYml.Spec().Stages[i].Actions[:j],
							pipelineYml.Spec().Stages[i].Actions[j+1:]...)
						break Exit
					} else {
						pipelineYml.Spec().Stages[i].Actions = pipelineYml.Spec().Stages[i].Actions[:j]
						break Exit
					}
				}
			}
		}

		if len(pipelineYml.Spec().Stages) == i+1 {
			break
		}
	}

	// 将tag插入release.yml的环境变量RELEASE_TAG
	if pipelineYml.Spec().Envs == nil {
		env := make(map[string]string)
		env["RELEASE_TAG"] = branchName
		pipelineYml.Spec().Envs = env
	} else {
		pipelineYml.Spec().Envs["RELEASE_TAG"] = branchName
	}

	newPipelineYml, err := pipelineyml.GenerateYml(pipelineYml.Spec())
	if err != nil {
		return "", err
	}

	if !isExistRelease {
		return "", errors.Errorf("pipeline.yml not exit release action")
	}

	return string(newPipelineYml), nil
}

func (p *Pipeline) AppCombos(appID uint64, req *spec.PipelineCombosReq) ([]apistructs.PipelineInvokedCombo, error) {
	// get pipelines
	pipelineReq := apistructs.PipelinePageListRequest{
		PageNum:       1,
		PageSize:      1000,
		LargePageSize: true,
		AllSources:    true,
		MustMatchLabelsQueryParams: []string{fmt.Sprintf("%s=%s", apistructs.LabelAppID,
			strconv.FormatUint(appID, 10))},
	}

	pipelinesResp, err := p.bdl.PageListPipeline(pipelineReq)
	if err != nil {
		return nil, apierrors.ErrGetPipeline.InternalError(err)
	}

	result := make([]apistructs.PipelineInvokedCombo, 0)
	// 将 pipelineYmlName 有关联的 combo 进行合并
	// 特殊处理 pipelineYmlName
	// pipeline.yml -> 1/PROD/master/pipeline.yml
	m := make(map[string]apistructs.PagePipeline)
	for i := range pipelinesResp.Pipelines {
		p := pipelinesResp.Pipelines[i]
		generateV1UniqueYmlName := GenerateV1UniquePipelineYmlName(p.Source, p.YmlName,
			p.FilterLabels[apistructs.LabelAppID], p.FilterLabels[apistructs.LabelBranch], p.Extra.DiceWorkspace)
		exist, ok := m[generateV1UniqueYmlName]
		// 取流水线 ID 最大的
		if !ok || p.ID > exist.ID {
			m[GenerateV1UniquePipelineYmlName(p.Source, p.YmlName, p.FilterLabels[apistructs.LabelAppID],
				p.FilterLabels[apistructs.LabelBranch], p.Extra.DiceWorkspace)] = p
		}
	}
	for ymlName, p := range m {
		ymlNameMap := map[string]struct{}{
			ymlName:                   {},
			p.YmlName:                 {},
			p.Extra.PipelineYmlNameV1: {},
			DecodeV1UniquePipelineYmlName(&p, ymlName): {},
		}
		// 保存需要聚合在一起的 ymlNames
		ymlNames := make([]string, 0)
		// 保存最短的 ymlName 用于 UI 展示
		shortYmlName := p.YmlName
		for name := range ymlNameMap {
			if name == "" {
				continue
			}
			if len(name) < len(shortYmlName) {
				shortYmlName = name
			}
			ymlNames = append(ymlNames, name)
		}
		result = append(result, apistructs.PipelineInvokedCombo{
			Branch: p.FilterLabels[apistructs.LabelBranch], Source: string(p.Source), YmlName: shortYmlName,
			PagingYmlNames: ymlNames, PipelineID: p.ID, Commit: p.Commit, Status: string(p.Status),
			TimeCreated: p.TimeCreated, CancelUser: p.Extra.CancelUser,
			TriggerMode: p.TriggerMode,
			Workspace:   p.Extra.DiceWorkspace,
		})
	}
	// 排序 ID DESC
	sort.Slice(result, func(i, j int) bool {
		return result[i].PipelineID > result[j].PipelineID
	})

	return result, nil
}

func (p *Pipeline) AllValidBranchWorkspaces(appID uint64) ([]apistructs.ValidBranch, error) {
	return p.bdl.GetAllValidBranchWorkspace(appID)
}

func (p *Pipeline) ConvertPipelineToV2(pv1 *apistructs.PipelineCreateRequest) (*apistructs.PipelineCreateRequestV2, error) {
	pv2 := &apistructs.PipelineCreateRequestV2{
		PipelineSource: apistructs.PipelineSourceDice,
		AutoRunAtOnce:  pv1.AutoRun,
		IdentityInfo:   apistructs.IdentityInfo{UserID: pv1.UserID},
	}

	labels := make(map[string]string, 0)
	// get app info
	app, err := p.bdl.GetApp(pv1.AppID)
	if err != nil {
		return nil, apierrors.ErrGetApp.InternalError(err)
	}

	// get newest commit info
	commit, err := p.bdl.GetGittarCommit(app.GitRepoAbbrev, pv1.Branch)
	if err != nil {
		return nil, apierrors.ErrGetGittarCommit.InternalError(err)
	}

	detail := apistructs.CommitDetail{
		CommitID: commit.ID,
		Repo:     app.GitRepo,
		RepoAbbr: app.GitRepoAbbrev,
		Author:   commit.Committer.Name,
		Email:    commit.Committer.Email,
		Time:     &commit.Committer.When,
		Comment:  commit.CommitMessage,
	}
	commitInfo, err := json.Marshal(&detail)
	if err != nil {
		return nil, err
	}
	labels[apistructs.LabelCommitDetail] = string(commitInfo)

	// 从 gittar 获取 pipeline.yml
	pipelineYmlName := pv1.PipelineYmlName
	if pipelineYmlName == "" {
		pipelineYmlName = apistructs.DefaultPipelineYmlName
	}

	strPipelineYml := pv1.PipelineYmlContent
	if strPipelineYml == "" {
		strPipelineYml, err = p.FetchPipelineYml(app.GitRepo, pv1.Branch, pipelineYmlName)
		if err != nil {
			return nil, apierrors.ErrGetGittarRepoFile.InternalError(err)
		}
	}

	pv2.PipelineYml = strPipelineYml
	rules, err := p.branchRuleSvc.Query(apistructs.ProjectScope, int64(app.ProjectID))
	if err != nil {
		return nil, apierrors.ErrGetGittarRepoFile.InternalError(err)
	}
	validBranch := diceworkspace.GetValidBranchByGitReference(pv1.Branch, rules)
	workspace := validBranch.Workspace

	// 塞入 publisher namespace, publisher 级别配置优先级低于用户指定
	relationResp, err := p.bdl.GetAppPublishItemRelationsGroupByENV(pv1.AppID)
	if err == nil && relationResp != nil {
		// 四个环境 publisherID 相同

		if publishItem, ok := relationResp.Data[strings.ToUpper(workspace)]; ok {
			pv2.ConfigManageNamespaces = append(pv2.ConfigManageNamespaces, publishItem.PublishItemNs...)
			// 根据 publishierID 获取 namespaces
			publisher, err := p.publisherSvc.Get(publishItem.PublisherID)
			if err == nil || publisher != nil {
				pv2.ConfigManageNamespaces = append(pv2.ConfigManageNamespaces, publisher.PipelineCmNamespaces...)
			}
		}
	}

	// make config namespace
	ns, err := p.makeNamespace(pv1.AppID, pv1.Branch, validBranch.Workspace)
	if err != nil {
		return nil, apierrors.ErrMakeConfigNamespace.InternalError(err)
	}
	pv2.ConfigManageNamespaces = append(pv2.ConfigManageNamespaces, ns...)

	// label
	labels[apistructs.LabelDiceWorkspace] = workspace
	labels[apistructs.LabelBranch] = pv1.Branch
	labels[apistructs.LabelOrgID] = strconv.FormatUint(app.OrgID, 10)
	labels[apistructs.LabelProjectID] = strconv.FormatUint(app.ProjectID, 10)
	labels[apistructs.LabelAppID] = strconv.FormatUint(app.ID, 10)

	pv2.Labels = labels

	// normalLabel
	normalLabels := make(map[string]string, 0)
	normalLabels[apistructs.LabelAppName] = app.Name
	normalLabels[apistructs.LabelProjectName] = app.ProjectName
	normalLabels[apistructs.LabelOrgName] = app.OrgName

	pv2.NormalLabels = normalLabels

	// clusterName
	pj, err := p.bdl.GetProject(app.ProjectID)
	if err != nil {
		return nil, apierrors.ErrGetProject.InternalError(err)
	}

	for ws, clusterName := range pj.ClusterConfig {
		if strutil.Equal(ws, workspace, true) {
			pv2.ClusterName = clusterName
			break
		}
	}

	// generate pipeline yaml name
	pv2.PipelineYmlName = GenerateV1UniquePipelineYmlName(pv2.PipelineSource, pipelineYmlName,
		strconv.FormatUint(app.ID, 10), pv1.Branch, workspace)

	return pv2, nil
}

func (p *Pipeline) makeNamespace(appID uint64, branch string, workspace string) ([]string, error) {
	ns, err := p.generatorPipelineNS(appID, branch, workspace)
	if err != nil {
		return nil, err
	}

	ws, err := generatorWorkspaceNS(appID, workspace)
	if err != nil {
		return nil, err
	}
	ns = append(ns, ws...)

	return ns, err
}

func generatorWorkspaceNS(appID uint64, workspace string) ([]string, error) {
	wsList := []string{
		fmt.Sprintf("app-%d-%s", appID, strings.ToLower(string(apistructs.DefaultWorkspace))),
	}
	wsList = append(wsList, fmt.Sprintf("app-%d-%s", appID, strings.ToLower(workspace)))

	return wsList, nil
}

func (p *Pipeline) generatorPipelineNS(appID uint64, branch string, workspace string) ([]string, error) {
	var cmNamespaces []string
	// 创建 default namespace
	cmNamespaces = append(cmNamespaces, fmt.Sprintf("%s-%d-default", cms.PipelineAppConfigNameSpacePrefix, appID))

	// TODO 直接使用workspace，不用映射 support hotfix
	// hotfix support 兼容判断,如果有历史遗留参数,使用历史分支级配置 不用workspace
	if gitflowutil.IsHotfix(branch) || gitflowutil.IsSupport(branch) {
		branchPrefix, _ := gitflowutil.GetReferencePrefix(branch)
		ns := fmt.Sprintf("%s-%d-%s", cms.PipelineAppConfigNameSpacePrefix, appID, branchPrefix)
		configs, err := p.cms.GetCmsNsConfigs(utils.WithInternalClientContext(context.Background()),
			&cmspb.CmsNsConfigsGetRequest{
				Ns:             ns,
				PipelineSource: apistructs.PipelineSourceDice.String(),
			})
		if err == nil {
			if len(configs.Data) > 0 {
				cmNamespaces = append(cmNamespaces, ns)
			}
		}
	} else {
		workspaceConfig := map[string]string{
			"PROD":    "master",
			"STAGING": "release",
			"TEST":    "develop",
			"DEV":     "feature",
		}
		// 创建 branch namespace
		pipelineNs, ok := workspaceConfig[workspace]
		if ok {
			cmNamespaces = append(cmNamespaces, fmt.Sprintf("%s-%d-%s", cms.PipelineAppConfigNameSpacePrefix, appID, pipelineNs))
		}
	}
	return cmNamespaces, nil
}

// GenerateV1UniquePipelineYmlName 为 v1 pipeline 返回 pipelineYmlName，该 name 在 source 下唯一
// 生成规则: AppID/DiceWorkspace/Branch/PipelineYmlPath
// 1) 100/PROD/master/ec/dws/itm/workflow/item_1d_df_process.workflow
// 2) 200/DEV/feature/dice/pipeline.yml
func GenerateV1UniquePipelineYmlName(source apistructs.PipelineSource, oriYmlName, appID, branch, workspace string) string {
	// source != (dice || bigdata) 时无需转换
	if !(source == apistructs.PipelineSourceDice || source == apistructs.PipelineSourceBigData) {
		return oriYmlName
	}
	// 若 originPipelineYmlPath 已经符合生成规则，则直接返回
	ss := strutil.Split(oriYmlName, "/", true)
	if len(ss) > 3 {
		oriAppID, _ := strconv.ParseUint(ss[0], 10, 64)
		_workspace := ss[1]
		branchWithYmlName := strutil.Join(ss[2:], "/", true)
		branchPrefix := strutil.Concat(branch, "/")
		if strconv.FormatUint(oriAppID, 10) == appID &&
			_workspace == workspace &&
			strutil.HasPrefixes(branchWithYmlName, branchPrefix) &&
			len(branchWithYmlName) > len(branchPrefix) {
			return oriYmlName
		}
	}
	return fmt.Sprintf("%s/%s/%s/%s", appID, workspace, branch, oriYmlName)
}

// DecodeV1UniquePipelineYmlName 根据 GenerateV1UniquePipelineYmlName 生成规则，反解析得到 originName
func DecodeV1UniquePipelineYmlName(p *apistructs.PagePipeline, name string) string {
	prefix := fmt.Sprintf("%s/%s/%s/", p.FilterLabels[apistructs.LabelAppID], p.Extra.DiceWorkspace,
		p.FilterLabels[apistructs.LabelBranch])
	return strutil.TrimPrefixes(name, prefix)
}

// PipelineCronUpdate pipeline cron update
func (p *Pipeline) PipelineCronUpdate(req apistructs.GittarPushPayloadEvent) error {
	appID, err := strconv.ParseInt(req.ApplicationID, 10, 64)
	if err != nil {
		return err
	}
	appDto, err := p.bdl.GetApp(uint64(appID))
	if err != nil {
		return err
	}
	branch := getBranch(req.Content.Ref)

	// get diffs between two commits
	compare, err := p.bdl.GetGittarCompare(req.Content.After, req.Content.Before, appID)
	if err != nil {
		return err
	}
	for _, v := range compare.Diff.Files {
		// is pipeline.yml rename to others,need to delete cron and stop it if cron enable
		if isPipelineYmlPath(v.OldName) && !isPipelineYmlPath(v.Name) {
			cron, err := p.GetPipelineCron(int64(appDto.ProjectID), appID, v.OldName, branch)
			if err != nil {
				logrus.Errorf("fail to GetPipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
				continue
			}
			if *cron.Enable {
				_, err = p.bdl.StopPipelineCron(cron.ID)
				if err != nil {
					logrus.Errorf("fail to StopPipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					continue
				}
			}
		}
		if isPipelineYmlPath(v.Name) {
			// if pipeline cron is not exist,no need to do anything
			cron, err := p.GetPipelineCron(int64(appDto.ProjectID), appID, v.OldName, branch)
			if err != nil {
				logrus.Errorf("fail to GetPipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
				continue
			}

			// if type is delete,need to delete cron and stop it if cron enable
			// if type is rename,need to delete cron and stop it if cron enable
			if v.Type == "delete" || v.Type == "rename" {
				if *cron.Enable {
					_, err = p.bdl.StopPipelineCron(cron.ID)
					if err != nil {
						logrus.Errorf("fail to StopPipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					}
				}
				continue
			}

			// if type modified, need to update cron and stop it if cron enable and cronExpr is empty
			if v.Type == "modified" {
				// get pipeline yml file content
				searchINode := appDto.ProjectName + "/" + appDto.Name + "/blob/" + branch + "/" + v.Name
				pipelineYml, err := p.bdl.GetGittarBlobNode("/wb/"+searchINode, req.OrgID)
				if err != nil {
					logrus.Errorf("fail to GetGittarBlobNode,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					continue
				}
				// get cronExpr from pipelineYml
				cronExpr, err := getCronExpr(pipelineYml)
				if err != nil {
					logrus.Errorf("fail to getCronExpr,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					continue
				}

				if err := p.bdl.UpdatePipelineCron(apistructs.PipelineCronUpdateRequest{
					ID:          cron.ID,
					PipelineYml: pipelineYml,
					CronExpr:    cronExpr,
				}); err != nil {
					logrus.Errorf("fail to UpdatePipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					continue
				}
				if *cron.Enable && cronExpr == "" {
					_, err = p.bdl.StopPipelineCron(cron.ID)
					if err != nil {
						logrus.Errorf("fail to StopPipelineCron,err: %s,path: %s,oldPath: %s", err.Error(), v.Name, v.OldName)
					}
				}
			}
		}
	}
	return nil
}

func getCronExpr(pipelineYmlStr string) (string, error) {
	if pipelineYmlStr == "" {
		return "", nil
	}
	pipelineYml, err := pipelineyml.New([]byte(pipelineYmlStr))
	if err != nil {
		return "", err
	}
	return pipelineYml.Spec().Cron, nil
}

func getBranch(ref string) string {
	var branchPrefix = "refs/heads/"
	if len(ref) <= len(branchPrefix) {
		return ""
	}
	return ref[len(branchPrefix):]
}

func isPipelineYmlPath(path string) bool {
	const pipelineYmlPathPattern = `^pipeline\.yml$|^\.dice/pipelines/.+\.yml$`
	matched, err := regexp.MatchString(pipelineYmlPathPattern, path)
	if err != nil {
		return false
	}
	return matched
}

// GetPipelineCron get pipeline cron
func (p *Pipeline) GetPipelineCron(projectID, appID int64, pathOld, branch string) (*apistructs.PipelineCronDTO, error) {
	workspace, err := p.getWorkSpace(projectID, branch)
	if err != nil {
		return nil, err
	}
	pipelineYmlNameOld := getPipelineYmlName(appID, workspace, branch, pathOld)
	pagingReq := apistructs.PipelineCronPagingRequest{
		AllSources: false,
		Sources:    []apistructs.PipelineSource{"dice"},
		YmlNames:   []string{pipelineYmlNameOld},
		PageSize:   1,
		PageNo:     1,
	}
	crons, err := p.bdl.PageListPipelineCrons(pagingReq)
	if err != nil {
		return nil, err
	}
	if len(crons.Data) == 0 {
		return nil, fmt.Errorf("the pipeline cron is not exist,pipelineName: %s", pipelineYmlNameOld)
	}
	return crons.Data[0], nil
}

// GetPipelineYmlName return PipelineYmlName eg: 63/TEST/develop/pipeline.yml
func getPipelineYmlName(appID int64, workspace, branch, path string) string {
	return strutil.Concat(strconv.FormatInt(appID, 10), "/", workspace, "/", branch, "/", path)
}

// GetWorkSpace return workSpace of project's workspaceConfig by given branch
func (p *Pipeline) getWorkSpace(project int64, branch string) (string, error) {
	rules, err := p.branchRuleSvc.Query(apistructs.ProjectScope, project)
	if err != nil {
		return "", err
	}
	branchRule := diceworkspace.GetValidBranchByGitReference(branch, rules)
	return branchRule.Workspace, nil
}
