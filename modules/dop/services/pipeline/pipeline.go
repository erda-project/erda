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

// Package pipeline pipeline相关的结构信息
package pipeline

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/branchrule"
	"github.com/erda-project/erda/modules/dop/services/publisher"
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
	cmNamespaces = append(cmNamespaces, fmt.Sprintf("%s-%d-default", apistructs.PipelineAppConfigNameSpacePreFix, appID))

	// TODO 直接使用workspace，不用映射 support hotfix
	// hotfix support 兼容判断,如果有历史遗留参数,使用历史分支级配置 不用workspace
	if gitflowutil.IsHotfix(branch) || gitflowutil.IsSupport(branch) {
		branchPrefix, _ := gitflowutil.GetReferencePrefix(branch)
		ns := fmt.Sprintf("%s-%d-%s", apistructs.PipelineAppConfigNameSpacePreFix, appID, branchPrefix)
		configs, err := p.bdl.GetPipelineCmsNsConfigs(ns, apistructs.PipelineCmsGetConfigsRequest{
			PipelineSource: "dice",
		})
		if err == nil {
			if len(configs) > 0 {
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
			cmNamespaces = append(cmNamespaces, fmt.Sprintf("%s-%d-%s", apistructs.PipelineAppConfigNameSpacePreFix, appID, pipelineNs))
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
