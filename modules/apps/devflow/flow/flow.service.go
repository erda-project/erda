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

package flow

import (
	context "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
	issuerelationpb "github.com/erda-project/erda-proto-go/apps/devflow/issuerelation/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/branchrule"
	"github.com/erda-project/erda/modules/dop/services/filetree"
	"github.com/erda-project/erda/modules/dop/services/permission"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type Service struct {
	p             *provider
	fileTree      *filetree.GittarFileTree
	branchRuleSvc *branchrule.BranchRule
	permission    *permission.Permission
}

func (p *Service) WithGittarFileTree(fileTree *filetree.GittarFileTree) {
	if p == nil {
		return
	}
	p.fileTree = fileTree
}

func (p *Service) WithBranchRule(branchRuleSvc *branchrule.BranchRule) {
	if p == nil {
		return
	}
	p.branchRuleSvc = branchRuleSvc
}

func (p *Service) WithPermission(permission *permission.Permission) {
	if p == nil {
		return
	}
	p.permission = permission
}

const issueRelationType = "devflow"
const gittarPrefixOpenApi = "/wb/"

const mrOpenState = "open"
const mrMergedState = "merged"

func (s *Service) CreateFlowNode(ctx context.Context, req *pb.CreateFlowNodeRequest) (*pb.CreateFlowNodeResponse, error) {
	app, err := s.p.bdl.GetApp(req.AppID)
	if err != nil {
		return nil, err
	}

	var respPath = gittarPrefixOpenApi + app.ProjectName + "/" + app.Name
	branches, err := s.p.bdl.GetGittarBranchesV2(respPath, apis.GetOrgID(ctx), true, apis.GetUserID(ctx))
	if err != nil {
		return nil, err
	}

	var findSourceBranch = false
	for _, branch := range branches {
		if branch == req.SourceBranch {
			findSourceBranch = true
		}
	}

	// auto create sourceBranch
	if !findSourceBranch {
		err := s.p.bdl.CreateGittarBranch(respPath, apistructs.GittarCreateBranchRequest{
			Name: req.SourceBranch,
			Ref:  req.TargetBranch,
		}, apis.GetOrgID(ctx), apis.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
	}

	// find branch merge
	result, err := s.p.bdl.ListMergeRequest(req.AppID, apis.GetUserID(ctx), apistructs.GittarQueryMrRequest{
		TargetBranch: req.TargetBranch,
		SourceBranch: req.SourceBranch,
		Page:         1,
		Size:         999,
	})
	if err != nil {
		return nil, err
	}
	mergeInfoList := result.List

	var mergeResult *apistructs.MergeRequestInfo
	for _, merge := range mergeInfoList {
		if merge.State == mrOpenState || merge.State == mrMergedState {
			mergeResult = merge
			break
		}
	}

	var mergeID int64
	var repoMergeID int
	if mergeResult == nil {
		// auto create merge
		mergeInfo, err := s.p.bdl.CreateMergeRequest(req.AppID, apis.GetUserID(ctx), apistructs.GittarCreateMergeRequest{
			TargetBranch:       req.TargetBranch,
			SourceBranch:       req.SourceBranch,
			Title:              "auto create devflow merge",
			Description:        "auto create devflow merge",
			RemoveSourceBranch: false,
			AssigneeID:         apis.GetUserID(ctx),
		})
		if err != nil {
			return nil, err
		}
		mergeID = mergeInfo.Id
		repoMergeID = mergeInfo.RepoMergeId
	} else {
		mergeID = mergeResult.Id
		repoMergeID = mergeResult.RepoMergeId
	}

	var extra = pb.IssueRelationExtra{
		AppID:       req.AppID,
		RepoMergeID: uint64(repoMergeID),
	}

	// create relation
	_, err = s.p.IssueRelation.Create(ctx, &issuerelationpb.CreateIssueRelationRequest{
		Type:     issueRelationType,
		IssueID:  req.IssueID,
		Relation: strconv.FormatInt(mergeID, 10),
		Extra:    jsonparse.JsonOneLine(&extra),
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateFlowNodeResponse{
		RepoMergeID: uint64(repoMergeID),
		MergeID:     uint64(mergeID),
	}, nil
}

func (s *Service) OperationDeploy(ctx context.Context, req *pb.OperationDeployRequest) (*pb.OperationDeployResponse, error) {
	if req.Enable == nil {
		return nil, fmt.Errorf("enable can not empty")
	}

	extra, err := s.getRelationExtra(ctx, req.MergeID)
	if err != nil {
		return nil, err
	}

	// todo get tempBranch from project rules
	//appDto, err := s.p.bdl.GetApp(extra.AppID)
	//if err != nil {
	//	return nil, err
	//}
	//result, err := s.p.FlowRule.GetDevFlowRulesByProjectID(ctx, &pb2.GetDevFlowRuleRequest{
	//	ProjectID: appDto.ProjectID,
	//})
	//if err != nil {
	//	return nil, err
	//}

	var tempBranch = ""
	var operation apistructs.MergeOperationTempBranchOperationType

	if req.Enable.Value {
		operation = apistructs.JoinToTempBranch
	} else {
		operation = apistructs.RemoveFromTempBranch
	}

	err = s.p.bdl.OperationTempBranch(extra.AppID, apis.GetUserID(ctx), apistructs.GittarMergeOperationTempBranchRequest{
		MergeID:    extra.RepoMergeID,
		Operation:  operation,
		TempBranch: tempBranch,
	})
	if err != nil {
		return nil, err
	}

	return &pb.OperationDeployResponse{}, nil
}

func (s *Service) DeleteFlowNode(ctx context.Context, req *pb.DeleteFlowNodeRequest) (*pb.DeleteFlowNodeResponse, error) {
	if req.MergeID <= 0 {
		return nil, fmt.Errorf("mergeID can not empty")
	}

	if req.IssueID <= 0 {
		return nil, fmt.Errorf("issueID can not empty")
	}

	issueRelations, err := s.p.IssueRelation.List(ctx, &issuerelationpb.ListIssueRelationRequest{
		Type: issueRelationType,
		Relations: []string{
			strconv.FormatUint(req.MergeID, 10),
		},
		IssueIDs: []uint64{req.IssueID},
	})
	if err != nil {
		return nil, err
	}

	data := issueRelations.Data
	var relationID string
	for _, relation := range data {
		if relation.IssueID == req.IssueID && relation.Relation == strconv.FormatUint(req.MergeID, 10) {
			relationID = relation.ID
		}
	}

	if relationID == "" {
		return &pb.DeleteFlowNodeResponse{}, nil
	}

	_, err = s.p.IssueRelation.Delete(ctx, &issuerelationpb.DeleteIssueRelationRequest{
		RelationID: relationID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.DeleteFlowNodeResponse{}, nil
}

// todo impl Reconstruction
func (s *Service) Reconstruction(ctx context.Context, req *pb.ReconstructionRequest) (*pb.ReconstructionResponse, error) {

	return &pb.ReconstructionResponse{}, nil
}

func (s *Service) buildRelationQueryRequest(req *pb.GetDevFlowInfoRequest) (*issuerelationpb.ListIssueRelationRequest, error) {
	query := &issuerelationpb.ListIssueRelationRequest{
		Type: issueRelationType,
	}
	if req.IssueID > 0 {
		query.IssueIDs = append(query.IssueIDs, req.IssueID)
	}
	if req.MergeID > 0 {
		query.Relations = append(query.Relations, strconv.FormatUint(req.MergeID, 10))
	}
	if len(query.Relations) == 0 && len(query.IssueIDs) == 0 {
		return nil, fmt.Errorf("relation and issue cannot be empty at the same time")
	}
	if req.ProjectID == 0 {
		return nil, fmt.Errorf("projectID can not empty")
	}
	return query, nil
}

func (s *Service) GetDevFlowInfo(ctx context.Context, req *pb.GetDevFlowInfoRequest) (*pb.GetDevFlowInfoResponse, error) {
	query, err := s.buildRelationQueryRequest(req)
	if err != nil {
		return nil, err
	}

	result, err := s.p.IssueRelation.List(ctx, query)
	if err != nil {
		return nil, err
	}
	relations := result.Data
	if len(relations) == 0 {
		return &pb.GetDevFlowInfoResponse{
			Status: getStatusDevFlowInfos(nil),
		}, nil
	}

	rules, err := s.branchRuleSvc.Query(apistructs.ProjectScope, int64(req.ProjectID))
	if err != nil {
		return nil, apierrors.ErrGetGittarRepoFile.InternalError(err)
	}
	// todo get tempBranch from project rules
	var tempBranch = ""

	var appInfoMap = map[uint64]apistructs.ApplicationDTO{}
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, err
	}
	appListResult, err := s.p.bdl.GetAppsByProjectSimple(req.ProjectID, uint64(orgID), apis.GetUserID(ctx))
	if err != nil {
		return nil, err
	}
	for _, app := range appListResult.List {
		appInfoMap[app.ID] = app
	}

	var relationExtraMap = map[string]*pb.IssueRelationExtra{}
	var relationMap = map[string]*issuerelationpb.IssueRelation{}
	var relationMergeInfoMap = map[string]*apistructs.MergeRequestInfo{}
	var needQueryTempBranchAppMap = map[uint64]uint64{}

	var devFlowInfos []*pb.DevFlowInfo

	work := limit_sync_group.NewWorker(5)
	for index := range relations {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			index := i[0].(int)
			relation := relations[index]

			var extra pb.IssueRelationExtra
			err = json.Unmarshal([]byte(relation.Extra), &extra)
			if err != nil {
				return fmt.Errorf("unmarshal extra %v error %v", relation.Extra, err)
			}

			err := s.permission.CheckAppAction(apistructs.IdentityInfo{UserID: apis.GetUserID(ctx)}, extra.AppID, apistructs.GetAction)
			if err != nil {
				var devFlowInfo = &pb.DevFlowInfo{}
				devFlowInfo.DevFlowNode = &pb.DevFlowNode{
					RepoMergeID: extra.RepoMergeID,
					AppID:       extra.AppID,
					AppName:     appInfoMap[extra.AppID].Name,
					TempBranch:  tempBranch,
					IssueID:     relation.IssueID,
				}
				devFlowInfo.HasPermission = false

				locker.Lock()
				devFlowInfos = append(devFlowInfos, devFlowInfo)
				locker.Unlock()
				return nil
			}

			mrInfo, err := s.p.bdl.GetMergeRequestDetail(extra.AppID, apis.GetUserID(ctx), extra.RepoMergeID)
			if err != nil {
				return err
			}

			if mrInfo.State != mrMergedState && mrInfo.State != mrOpenState {
				go func() {
					_, err := s.p.IssueRelation.Delete(ctx, &issuerelationpb.DeleteIssueRelationRequest{
						RelationID: relation.ID,
					})
					if err != nil {
						s.p.Log.Errorf("delete not open mr relation error %v", err)
					}
				}()
				return nil
			}

			locker.Lock()
			defer locker.Unlock()

			relationMap[relation.ID] = relation
			relationExtraMap[relation.ID] = &extra
			relationMergeInfoMap[relation.ID] = mrInfo
			if mrInfo.IsJoinTempBranch() {
				needQueryTempBranchAppMap[extra.AppID] = extra.AppID
			}
			return nil
		}, index)
	}
	if work.Do().Error() != nil {
		return nil, work.Error()
	}

	var appTempBranchCommitMap = map[uint64]string{}
	var appTempBranchChangeBranchListMap = map[uint64][]*pb.ChangeBranch{}

	for _, appID := range needQueryTempBranchAppMap {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			appID := i[0].(uint64)
			appDto := appInfoMap[appID]
			var resp = fmt.Sprintf(gittarPrefixOpenApi + appDto.ProjectName + "/" + appDto.Name)
			commit, err := s.p.bdl.ListGittarCommit(resp, tempBranch, apis.GetUserID(ctx), apis.GetOrgID(ctx))
			if err != nil {
				return err
			}

			locker.Lock()
			appTempBranchCommitMap[appID] = commit.ID
			// todo query branch change branch list
			appTempBranchChangeBranchListMap = map[uint64][]*pb.ChangeBranch{}
			locker.Unlock()
			return nil
		}, appID)
	}
	if work.Do().Error() != nil {
		return nil, work.Error()
	}

	for relationID := range relationMap {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			relationID := i[0].(string)
			relation := relationMap[relationID]

			extra := relationExtraMap[relationID]
			mrInfo := relationMergeInfoMap[relationID]
			appDto := appInfoMap[extra.AppID]

			var branch string
			var commitID string
			var changeBranches []*pb.ChangeBranch
			var infos []*pb.PipelineStepInfo

			if mrInfo.IsJoinTempBranch() {
				branch = tempBranch
				commitID = appTempBranchCommitMap[appDto.ID]
				changeBranches = appTempBranchChangeBranchListMap[appDto.ID]
			} else {
				branch = mrInfo.SourceBranch
				commitID = mrInfo.SourceSha
				changeBranches = append(changeBranches, &pb.ChangeBranch{
					Commit:      commitID,
					BranchName:  branch,
					RepoMergeID: extra.RepoMergeID,
				})
			}
			infos, err := s.listPipelineStepInfo(ctx, &appDto, branch, diceworkspace.GetValidBranchByGitReference(branch, rules).Workspace, commitID)
			if err != nil {
				return err
			}

			var devFlowInfo = &pb.DevFlowInfo{}
			devFlowInfo.ChangeBranch = changeBranches
			devFlowInfo.Commit = commitID
			devFlowInfo.PipelineStepInfos = infos
			devFlowInfo.HasPermission = true
			devFlowInfo.DevFlowNode = &pb.DevFlowNode{
				RepoMergeID:          extra.RepoMergeID,
				MergeID:              uint64(mrInfo.Id),
				AppID:                extra.AppID,
				AppName:              appInfoMap[extra.AppID].Name,
				TargetBranch:         mrInfo.TargetBranch,
				SourceBranch:         mrInfo.SourceBranch,
				JoinTempBranchStatus: mrInfo.JoinTempBranchStatus,
				TempBranch:           tempBranch,
				IssueID:              relation.IssueID,
				IsJoinTempBranch:     mrInfo.IsJoinTempBranch(),
			}

			locker.Lock()
			devFlowInfos = append(devFlowInfos, devFlowInfo)
			locker.Unlock()
			return nil
		}, relationID)
	}
	err = work.Do().Error()
	if err != nil {
		return nil, err
	}

	return &pb.GetDevFlowInfoResponse{
		DevFlowInfos: devFlowInfos,
		Status:       getStatusDevFlowInfos(devFlowInfos),
	}, nil
}

func (s *Service) queryAllPipelineYmlAndDoFunc(ctx context.Context, appDto *apistructs.ApplicationDTO, branch string,
	do func(ctx context.Context, locker *limit_sync_group.Locker, content string, node *apistructs.UnifiedFileTreeNode) error) error {

	ymlNodes, err := s.getAllPipelineYml(ctx, appDto, branch)
	if err != nil {
		return err
	}

	worker := limit_sync_group.NewWorker(5)
	for index := range ymlNodes {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			index := i[0].(int)
			node := ymlNodes[index]

			content, err := s.getPipelineYmlNodeContent(ctx, appDto, branch, node)
			if err != nil {
				return err
			}

			return do(ctx, locker, content, node)
		}, index)
	}
	return worker.Do().Error()
}

func (s *Service) listPipelineStepInfo(ctx context.Context, appDto *apistructs.ApplicationDTO, branch string, workSpace string, commit string) ([]*pb.PipelineStepInfo, error) {
	var infos []*pb.PipelineStepInfo
	err := s.queryAllPipelineYmlAndDoFunc(ctx, appDto, branch, func(ctx context.Context, locker *limit_sync_group.Locker, content string, node *apistructs.UnifiedFileTreeNode) error {
		pipelineYml, err := pipelineyml.New([]byte(content))
		if err != nil {
			s.p.Log.Errorf("failed to parse %v yaml:%v \n err:%v", content, pipelineYml, err)
			return nil
		}

		if !pipelineYml.HasOnPushBranch(branch) {
			return nil
		}

		inodeBytes, err := base64.URLEncoding.DecodeString(node.Inode)
		if err != nil {
			return err
		}

		var stepInfo = pb.PipelineStepInfo{
			YmlName: node.Name,
		}

		ymlName := getFullYmlName(string(inodeBytes), node.Name)
		result, err := s.p.bdl.PageListPipeline(apistructs.PipelinePageListRequest{
			PageNo:   1,
			PageSize: 1,
			Sources: []apistructs.PipelineSource{
				apistructs.PipelineSourceDice,
			},
			YmlNames: []string{
				fmt.Sprintf("%v/%v/%v/%v", appDto.ID, workSpace, branch, ymlName),
			},
			DescCols: []string{"id"},
		})
		if err != nil {
			s.p.Log.Errorf("failed to list ymlName %v branch %v pipeline info err: %v", ymlName, branch, err)
		} else {
			if len(result.Pipelines) > 0 && result.Pipelines[0].Commit == commit {
				stepInfo.PipelineID = result.Pipelines[0].ID
				stepInfo.Status = result.Pipelines[0].Status.String()
			}
		}

		locker.Lock()
		infos = append(infos, &stepInfo)
		locker.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func getFullYmlName(parseInode string, ymlName string) string {
	if ymlName != apistructs.DefaultPipelineYmlName {
		if strings.Contains(parseInode, apistructs.DicePipelinePath) {
			ymlName = apistructs.DicePipelinePath + "/" + ymlName
		} else {
			ymlName = apistructs.ErdaPipelinePath + "/" + ymlName
		}
	}
	return ymlName
}

func (s *Service) getPipelineYmlNodeContent(ctx context.Context, appDto *apistructs.ApplicationDTO, branch string, node *apistructs.UnifiedFileTreeNode) (string, error) {

	inodeBytes, err := base64.URLEncoding.DecodeString(node.Inode)
	if err != nil {
		return "", apierrors.ErrGetGittarFileTreeNode.InternalError(err)
	}

	ymlName := getFullYmlName(string(inodeBytes), node.Name)
	inode := fmt.Sprintf("%v/%v/blob/%v/%v", appDto.ProjectName, appDto.Name, branch, ymlName)

	content, err := s.p.bdl.GetGittarBlobNode(gittarPrefixOpenApi+inode, apis.GetOrgID(ctx), apis.GetUserID(ctx))
	if err != nil {
		return "", apierrors.ErrGetGittarFileTreeNode.InternalError(err)
	}
	return content, nil
}

func getStatusDevFlowInfos(devFlowInfos []*pb.DevFlowInfo) string {
	if len(devFlowInfos) == 0 {
		return "none"
	}

	var pipelineNotRunNum = 0
	var pipelineFailedNum = 0
	var pipelineSuccessNum = 0
	var pipelineNum = 0

	for _, info := range devFlowInfos {
		if info.DevFlowNode.JoinTempBranchStatus == apistructs.JoinTempBranchFailedStatus {
			return "mergeFailed"
		}
		for _, pipelineInfo := range info.PipelineStepInfos {
			if pipelineInfo.Status == "" {
				pipelineNotRunNum++
			}
			if apistructs.PipelineStatus(pipelineInfo.Status).IsFailedStatus() {
				pipelineFailedNum++
			}
			if apistructs.PipelineStatus(pipelineInfo.Status).IsSuccessStatus() {
				pipelineSuccessNum++
			}
		}
		pipelineNum++
	}

	if pipelineNotRunNum > 0 {
		return "pipelineNotRun"
	}
	if pipelineFailedNum > 0 {
		return "pipelineFiled"
	}
	if pipelineSuccessNum != pipelineNum {
		return "pipelineRunning"
	}
	return "success"
}

func (s *Service) getAllPipelineYml(ctx context.Context, app *apistructs.ApplicationDTO, branch string) ([]*apistructs.UnifiedFileTreeNode, error) {
	orgID, err := strconv.ParseInt(apis.GetOrgID(ctx), 10, 64)
	if err != nil {
		return nil, err
	}
	var nodes []*apistructs.UnifiedFileTreeNode

	worker := limit_sync_group.NewWorker(3)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var defaultYmlPath = fmt.Sprintf("%v/%v/tree/%v", app.ProjectID, app.ID, branch)
		pipelineTreeData, err := s.fileTree.ListFileTreeNodes(apistructs.UnifiedFileTreeNodeListRequest{
			Scope:   apistructs.FileTreeScopeProjectApp,
			ScopeID: strconv.FormatUint(app.ProjectID, 10),
			Pinode:  base64.URLEncoding.EncodeToString([]byte(defaultYmlPath)),
			IdentityInfo: apistructs.IdentityInfo{
				UserID: apis.GetUserID(ctx),
			},
		}, uint64(orgID))
		if err != nil {
			s.p.Log.Debug("failed to get %v path yml error %v", apistructs.DefaultPipelineYmlName, err)
			return nil
		}

		locker.Lock()
		defer locker.Unlock()
		for _, node := range pipelineTreeData {
			if node.Name == apistructs.DefaultPipelineYmlName {
				nodes = append(nodes, node)
				break
			}
		}
		return nil
	})

	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var diceYmlPath = fmt.Sprintf("%v/%v/tree/%v/%v", app.ProjectID, app.ID, branch, apistructs.DicePipelinePath)
		diceTreeData, err := s.fileTree.ListFileTreeNodes(apistructs.UnifiedFileTreeNodeListRequest{
			Scope:   apistructs.FileTreeScopeProjectApp,
			ScopeID: strconv.FormatUint(app.ProjectID, 10),
			Pinode:  base64.URLEncoding.EncodeToString([]byte(diceYmlPath)),
			IdentityInfo: apistructs.IdentityInfo{
				UserID: apis.GetUserID(ctx),
			},
		}, uint64(orgID))
		if err != nil {
			s.p.Log.Debug("failed to get %v path yml error %v", apistructs.DicePipelinePath, err)
			return nil
		}

		locker.Lock()
		defer locker.Unlock()
		nodes = append(nodes, diceTreeData...)
		return nil
	})

	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var erdaYmlPath = fmt.Sprintf("%v/%v/tree/%v/%v", app.ProjectID, app.ID, branch, apistructs.ErdaPipelinePath)
		erdaTreeData, err := s.fileTree.ListFileTreeNodes(apistructs.UnifiedFileTreeNodeListRequest{
			Scope:   apistructs.FileTreeScopeProjectApp,
			ScopeID: strconv.FormatUint(app.ProjectID, 10),
			Pinode:  base64.URLEncoding.EncodeToString([]byte(erdaYmlPath)),
			IdentityInfo: apistructs.IdentityInfo{
				UserID: apis.GetUserID(ctx),
			},
		}, uint64(orgID))
		if err != nil {
			s.p.Log.Debug("failed to get %v path yml error %v", apistructs.ErdaPipelinePath, err)
			return nil
		}

		locker.Lock()
		defer locker.Unlock()
		nodes = append(nodes, erdaTreeData...)
		return nil
	})
	err = worker.Do().Error()
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (s *Service) getRelationExtra(ctx context.Context, mergeID uint64) (*pb.IssueRelationExtra, error) {
	resp, err := s.p.IssueRelation.List(ctx, &issuerelationpb.ListIssueRelationRequest{
		Type: issueRelationType,
		Relations: []string{
			strconv.FormatInt(int64(mergeID), 10),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("not find this devflow merge id %v", mergeID)
	}

	var extra pb.IssueRelationExtra
	err = json.Unmarshal([]byte(resp.Data[0].Extra), &extra)
	if err != nil {
		return nil, fmt.Errorf("unmarshal extra %v error %v", resp.Data[0].Extra, err)
	}
	return &extra, nil
}
