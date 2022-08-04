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
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
	issuerelationpb "github.com/erda-project/erda-proto-go/apps/devflow/issuerelation/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	flowrulepb "github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	gittarpb "github.com/erda-project/erda-proto-go/openapiv1/gittar/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/devflow/flow/db"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/apps/dop/services/branchrule"
	"github.com/erda-project/erda/internal/apps/dop/services/filetree"
	"github.com/erda-project/erda/internal/apps/dop/services/permission"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/pbutil"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

const (
	MultiBranchBranchType  = "multi_branch"
	SingleBranchBranchType = "single_branch"
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

// CreateFlowNode currentBranch checkout from sourceBranch
func (s *Service) CreateFlowNode(ctx context.Context, req *pb.CreateFlowNodeRequest) (*pb.CreateFlowNodeResponse, error) {
	issue, err := s.p.Issue.GetIssue(int64(req.IssueID), &commonpb.IdentityInfo{
		UserID: apis.GetUserID(ctx),
	})
	if err != nil {
		return nil, err
	}

	app, err := s.p.bdl.GetApp(req.AppID)
	if err != nil {
		return nil, err
	}
	repoPath := makeGittarRepoPath(app)

	branchPolicy, err := s.findBranchPolicyByName(ctx, issue.ProjectID, req.FlowRuleName)
	if err != nil {
		return nil, err
	}
	if !isRefPatternMatch(req.CurrentBranch, branchPolicy.Branch) {
		return nil, fmt.Errorf("the current branch does not match the rule")
	}

	var sourceBranch string
	if branchPolicy.Policy != nil {
		sourceBranch = branchPolicy.Policy.SourceBranch
	}

	// Idempotent create currentBranch, code node
	if err = s.IdempotentCreateBranch(ctx, repoPath, sourceBranch, req.CurrentBranch); err != nil {
		return nil, err
	}

	flow := db.DevFlow{
		Scope: db.Scope{
			OrgID:   app.OrgID,
			OrgName: app.OrgName,
			AppID:   req.AppID,
			AppName: app.Name,
		},
		Operator: db.Operator{
			Creator: apis.GetUserID(ctx),
		},
		Branch:       req.CurrentBranch,
		IssueID:      req.IssueID,
		FlowRuleName: req.FlowRuleName,
	}
	if err = s.p.dbClient.CreateDevFlow(&flow); err != nil {
		return nil, err
	}

	return &pb.CreateFlowNodeResponse{Data: flow.Covert()}, nil
}

func (s *Service) getFlowRuleNameBranchPolicyMap(ctx context.Context, projectID uint64) (map[string]branchPolicy, error) {
	devFlowRule, err := s.p.DevFlowRule.GetDevFlowRulesByProjectID(ctx, &flowrulepb.GetDevFlowRuleRequest{ProjectID: projectID})
	if err != nil {
		return nil, err
	}
	flowRuleNameBranchPolicyMap := make(map[string]branchPolicy)
	for _, flow := range devFlowRule.Data.Flows {
		var branchPolicy branchPolicy
		for _, policy := range devFlowRule.Data.BranchPolicies {
			if policy.Branch == flow.TargetBranch {
				if policy.Policy != nil {
					branchPolicy.sourceBranch = policy.Policy.SourceBranch
					branchPolicy.tempBranch = policy.Policy.TempBranch
					if policy.Policy.TargetBranch != nil {
						branchPolicy.targetBranch = policy.Policy.TargetBranch.MergeRequest
					}
				}
			}
		}
		flowRuleNameBranchPolicyMap[flow.Name] = branchPolicy
	}
	return flowRuleNameBranchPolicyMap, nil
}

func (s *Service) findBranchPolicyByName(ctx context.Context, projectID uint64, flowRuleName string) (*flowrulepb.BranchPolicy, error) {
	devFlowRule, err := s.p.DevFlowRule.GetDevFlowRulesByProjectID(ctx, &flowrulepb.GetDevFlowRuleRequest{ProjectID: projectID})
	if err != nil {
		return nil, err
	}

	var findBranch string
	for _, v := range devFlowRule.Data.Flows {
		if v.Name == flowRuleName {
			findBranch = v.TargetBranch
		}
	}
	if findBranch == "" {
		return nil, fmt.Errorf("not found devFlowRule")
	}
	for _, v := range devFlowRule.Data.BranchPolicies {
		if v.Branch == findBranch {
			return v, nil
		}
	}
	return nil, fmt.Errorf("not found branchPolicy")
}

func (s *Service) makeGittarRepoPathByAppID(appID uint64) (string, error) {
	app, err := s.p.bdl.GetApp(appID)
	if err != nil {
		return "", err
	}
	return makeGittarRepoPath(app), nil
}

func makeGittarRepoPath(app *apistructs.ApplicationDTO) string {
	if app == nil {
		return ""
	}
	return filepath.Join(gittarPrefixOpenApi, app.ProjectName, app.Name)
}

func isRefPatternMatch(currentBranch string, branchRule string) bool {
	branchRules := strings.Split(branchRule, ",")
	return diceworkspace.IsRefPatternMatch(currentBranch, branchRules)
}

func (s *Service) makeMrTitle(ctx context.Context, branch string) string {
	return fmt.Sprintf("Automatic merge requests for %s", branch)
}

func (s *Service) makeMrDescByIssueID(ctx context.Context, issueID uint64) (string, error) {
	issue, err := s.p.Issue.GetIssue(int64(issueID), &commonpb.IdentityInfo{UserID: apis.GetUserID(ctx)})
	if err != nil {
		return "", err
	}
	return s.makeMrDesc(ctx, issue), nil
}

func (s *Service) makeMrDesc(ctx context.Context, issue *issuepb.Issue) string {
	if issue == nil {
		return ""
	}
	return fmt.Sprintf("%s: #%d %s", s.p.Trans.Text(apis.Language(ctx), issue.Type.String()), issue.Id, issue.Title)
}

func (s *Service) OperationMerge(ctx context.Context, req *pb.OperationMergeRequest) (*pb.OperationMergeResponse, error) {
	if req.Enable == nil {
		return nil, fmt.Errorf("enable can not empty")
	}

	devFlow, err := s.p.dbClient.GetDevFlow(req.DevFlowID)
	if err != nil {
		return nil, err
	}

	app, err := s.p.bdl.GetApp(devFlow.AppID)
	if err != nil {
		return nil, err
	}

	branchPolicy, err := s.findBranchPolicyByName(ctx, app.ProjectID, devFlow.FlowRuleName)
	if err != nil {
		return nil, err
	}

	var sourceBranch, tempBranch, targetBranch string
	if branchPolicy.Policy != nil {
		sourceBranch = branchPolicy.Policy.SourceBranch
		tempBranch = branchPolicy.Policy.TempBranch
		if branchPolicy.Policy.TargetBranch != nil {
			targetBranch = branchPolicy.Policy.TargetBranch.MergeRequest
		}
	}
	if tempBranch == "" {
		return nil, fmt.Errorf("tempBranch can not empty")
	}
	if sourceBranch == "" {
		return nil, fmt.Errorf("sourceBranch can not empty")
	}

	// Merged into tempBranch success will set isJoinTempBranch = true
	if !req.Enable.Value {
		devFlow.IsJoinTempBranch = req.Enable.Value
		if err = s.p.dbClient.UpdateDevFlow(devFlow); err != nil {
			return nil, err
		}
	}

	if req.Enable.Value {
		err = s.JoinTempBranch(ctx, tempBranch, sourceBranch, app, devFlow)
		if err != nil {
			return nil, err
		}
	} else {
		// Is already removed
		if !devFlow.IsJoinTempBranch {
			return &pb.OperationMergeResponse{}, nil
		}
		err = s.RejoinTempBranch(ctx, tempBranch, sourceBranch, targetBranch, devFlow, app)
		if err != nil {
			return nil, err
		}
	}

	userID := apis.GetUserID(ctx)
	if err := s.CreateFlowEvent(&CreateFlowRequest{
		ProjectID: app.ProjectID,
		AppID:     app.ID,
		OrgID:     app.OrgID,
		Data: &pb.FlowEventData{
			IssueID:          devFlow.IssueID,
			Operator:         userID,
			TempBranch:       tempBranch,
			SourceBranch:     sourceBranch,
			TargetBranch:     targetBranch,
			AppName:          app.DisplayName,
			IsJoinTempBranch: strconv.FormatBool(req.Enable.Value),
		},
	}); err != nil {
		s.p.Log.Errorf("failed to create flow event, err: %v", err)
	}

	return &pb.OperationMergeResponse{}, nil
}

type CreateFlowRequest struct {
	Data      *pb.FlowEventData
	ProjectID uint64
	AppID     uint64
	OrgID     uint64
}

func (s *Service) CreateFlowEvent(req *CreateFlowRequest) error {
	projectModel, err := s.p.bdl.GetProject(req.ProjectID)
	if err != nil {
		return err
	}

	issueModel, err := s.p.Issue.GetIssue(int64(req.Data.IssueID), &commonpb.IdentityInfo{UserID: req.Data.Operator})
	if err != nil {
		return err
	}
	req.Data.ProjectName = projectModel.Name
	req.Data.Params = make(map[string]*structpb.Value)
	req.Data.Params["title"] = structpb.NewStringValue(issueModel.Title)
	eventReq := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         "dev_flow",
			Action:        "update",
			ProjectID:     strconv.FormatUint(req.ProjectID, 10),
			ApplicationID: strconv.FormatUint(req.AppID, 10),
			OrgID:         strconv.FormatUint(req.OrgID, 10),
		},
		Sender:  bundle.SenderDOP,
		Content: req.Data,
	}
	return s.p.bdl.CreateEvent(eventReq)
}

func (s *Service) JoinTempBranch(ctx context.Context, tempBranch, sourceBranch string, appDto *apistructs.ApplicationDTO, devFlow *db.DevFlow) error {
	repoPath := makeGittarRepoPath(appDto)

	// Idempotent create tempBranch
	if err := s.IdempotentCreateBranch(ctx, repoPath, sourceBranch, tempBranch); err != nil {
		return err
	}

	// Merge sourceBranch to tempBranch
	return s.MergeToTempBranch(ctx, tempBranch, appDto.ID, devFlow)
}

func (s *Service) RejoinTempBranch(ctx context.Context, tempBranch, sourceBranch, targetBranch string, devFlow *db.DevFlow, app *apistructs.ApplicationDTO) error {
	devFlows, err := s.p.dbClient.ListDevFlowByFlowRuleName(devFlow.FlowRuleName)
	if err != nil {
		return err
	}

	repoPath := makeGittarRepoPath(app)
	// Delete tempBranch
	if err = s.IdempotentDeleteBranch(ctx, repoPath, tempBranch); err != nil {
		return err
	}

	// Idempotent create tempBranch
	if err = s.IdempotentCreateBranch(ctx, repoPath, sourceBranch, tempBranch); err != nil {
		return err
	}

	// Merge the branch to tempBranch
	for _, v := range devFlows {
		if !v.IsJoinTempBranch {
			continue
		}
		isMROpenedOrNotCreated, err := s.isMROpenedOrNotCreated(ctx, v.Branch, targetBranch, app.ID)
		if err != nil {
			return err
		}

		if !isMROpenedOrNotCreated {
			continue
		}
		if err = s.MergeToTempBranch(ctx, tempBranch, app.ID, &v); err != nil {
			s.p.Log.Errorf("failed to MergeWithBranch, err: %v", err)
		}
	}
	return nil
}

func (s *Service) isMROpenedOrNotCreated(ctx context.Context, currentBranch, targetBranch string, appID uint64) (bool, error) {
	result, err := s.p.bdl.ListMergeRequest(appID, apis.GetUserID(ctx), apistructs.GittarQueryMrRequest{
		TargetBranch: targetBranch,
		SourceBranch: currentBranch,
		Page:         1,
		Size:         999,
	})
	if err != nil {
		return false, err
	}
	// mr not created
	if result.Total == 0 {
		return true, nil
	}
	for _, v := range result.List {
		if v.State == mrOpenState {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) MergeToTempBranch(ctx context.Context, tempBranch string, appID uint64, devFlow *db.DevFlow) error {
	_, err := s.p.bdl.MergeWithBranch(apis.GetUserID(ctx), apistructs.GittarMergeWithBranchRequest{
		SourceBranch: devFlow.Branch,
		TargetBranch: tempBranch,
		AppID:        appID,
	})

	if err != nil {
		devFlow.JoinTempBranchStatus = apistructs.JoinTempBranchFailedStatus + err.Error()
	} else {
		devFlow.JoinTempBranchStatus = apistructs.JoinTempBranchSuccessStatus
		devFlow.IsJoinTempBranch = true
	}

	return s.p.dbClient.UpdateDevFlow(devFlow)
}

func (s *Service) IdempotentCreateBranch(ctx context.Context, repoPath, sourceBranch, newBranch string) error {
	if sourceBranch == "" {
		return fmt.Errorf("the sourceBranch can not be empty")
	}
	exists, err := s.JudgeBranchIsExists(ctx, repoPath, sourceBranch)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the sourceBranch not exists")
	}

	exists, err = s.JudgeBranchIsExists(ctx, repoPath, newBranch)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	err = s.p.bdl.CreateGittarBranch(repoPath, apistructs.GittarCreateBranchRequest{
		Name: newBranch,
		Ref:  sourceBranch,
	}, apis.GetOrgID(ctx), apis.GetUserID(ctx))
	return err
}

func (s *Service) IdempotentDeleteBranch(ctx context.Context, repoPath, branch string) error {
	exists, err := s.JudgeBranchIsExists(ctx, repoPath, branch)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	err = s.p.bdl.DeleteGittarBranch(repoPath, apis.GetOrgID(ctx), branch, apis.GetUserID(ctx))
	return err
}

func (s *Service) JudgeBranchIsExists(ctx context.Context, repoPath, branch string) (has bool, err error) {
	branches, err := s.p.bdl.GetGittarBranchesV2(repoPath, apis.GetOrgID(ctx), true, apis.GetUserID(ctx))
	if err != nil {
		return false, err
	}

	for _, v := range branches {
		if v == branch {
			return true, nil
		}
	}
	return false, nil
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

// Reconstruction todo impl
func (s *Service) Reconstruction(ctx context.Context, req *pb.ReconstructionRequest) (*pb.ReconstructionResponse, error) {

	return &pb.ReconstructionResponse{}, nil
}

type branchPolicy struct {
	tempBranch   string
	targetBranch string
	sourceBranch string
}

func (s *Service) GetDevFlowInfo(ctx context.Context, req *pb.GetDevFlowInfoRequest) (*pb.GetDevFlowInfoResponse, error) {
	devFlows, err := s.listDevFlowByReq(req)
	if err != nil {
		return nil, err
	}
	if len(devFlows) == 0 {
		return &pb.GetDevFlowInfoResponse{}, nil
	}
	issueID := devFlows[0].IssueID

	issue, err := s.p.Issue.GetIssue(int64(issueID), &commonpb.IdentityInfo{UserID: apis.GetUserID(ctx)})
	if err != nil {
		return nil, err
	}

	ruleNameBranchPolicyMap, err := s.getFlowRuleNameBranchPolicyMap(ctx, issue.ProjectID)
	if err != nil {
		return nil, err
	}

	rules, err := s.branchRuleSvc.Query(apistructs.ProjectScope, int64(issue.ProjectID))
	if err != nil {
		return nil, err
	}

	appMap := make(map[uint64]*apistructs.ApplicationDTO)
	appTempBranchChangeBranchListMap := make(map[string][]*pb.ChangeBranch)
	appTempBranchCommitMap := make(map[string]*apistructs.Commit)
	devFlowInfos := make([]*pb.DevFlowInfo, 0, len(devFlows))

	work := limit_sync_group.NewWorker(5)
	for index := range devFlows {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			index := i[0].(int)
			devFlow := devFlows[index]

			devFlowInfo := pb.DevFlowInfo{
				DevFlow:       devFlow.Covert(),
				HasPermission: true,
			}
			// Check app permission
			err = s.permission.CheckAppAction(apistructs.IdentityInfo{UserID: apis.GetUserID(ctx)}, devFlow.AppID, apistructs.GetAction)
			if err != nil {
				devFlowInfo.HasPermission = false
				locker.Lock()
				devFlowInfos = append(devFlowInfos, &devFlowInfo)
				locker.Unlock()
				return nil
			}

			currentBranch := devFlow.Branch
			tempBranch := ruleNameBranchPolicyMap[devFlow.FlowRuleName].tempBranch
			sourceBranch := ruleNameBranchPolicyMap[devFlow.FlowRuleName].sourceBranch
			targetBranch := ruleNameBranchPolicyMap[devFlow.FlowRuleName].targetBranch

			var app *apistructs.ApplicationDTO
			locker.Lock()
			if _, ok := appMap[devFlow.AppID]; !ok {
				app, err = s.p.bdl.GetApp(devFlow.AppID)
				if err != nil {
					return err
				}
				appMap[app.ID] = app
			} else {
				app = appMap[devFlow.AppID]
			}
			locker.Unlock()
			repoPath := makeGittarRepoPath(app)

			currentBranchExists, err := s.JudgeBranchIsExists(ctx, repoPath, currentBranch)
			if err != nil {
				return err
			}
			var currentBranchCommit, baseCommit *apistructs.Commit
			if currentBranchExists {
				branchDetail, err := s.p.bdl.GetGittarBranchDetail(repoPath, apis.GetOrgID(ctx), currentBranch, apis.GetUserID(ctx))
				if err != nil {
					return err
				}
				currentBranchCommit = branchDetail.Commit
			}

			if tempBranch != "" {
				if err = s.IdempotentCreateBranch(ctx, repoPath, sourceBranch, tempBranch); err != nil {
					return err
				}

				locker.Lock()
				if _, ok := appTempBranchCommitMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)]; !ok {
					var commit *apistructs.Commit
					if tempBranch != "" {
						commit, err = s.p.bdl.ListGittarCommit(repoPath, tempBranch, apis.GetUserID(ctx), apis.GetOrgID(ctx))
						if err != nil {
							return err
						}
					}
					appTempBranchCommitMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)] = commit
				}

				if devFlow.IsJoinTempBranch && currentBranchExists {
					baseCommit, err = s.p.bdl.GetMergeBase(apis.GetUserID(ctx), apistructs.GittarMergeBaseRequest{
						SourceBranch: currentBranch,
						TargetBranch: tempBranch,
						AppID:        devFlow.AppID,
					})
					if err != nil {
						return err
					}
					appTempBranchChangeBranchListMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)] =
						append(appTempBranchChangeBranchListMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)], &pb.ChangeBranch{
							Commit:     commitConvert(baseCommit),
							BranchName: currentBranch,
						})
				}
				locker.Unlock()
			}

			var pipelineStepInfo []*pb.PipelineStepInfo
			if devFlow.IsJoinTempBranch {
				var commitID string
				if appTempBranchCommitMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)] != nil {
					commitID = appTempBranchCommitMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)].ID
				}
				if tempBranch != "" && commitID != "" {
					pipelineStepInfo, err = s.listPipelineStepInfo(ctx, app, tempBranch, diceworkspace.GetValidBranchByGitReference(tempBranch, rules).Workspace, commitID)
					if err != nil {
						return err
					}
				}
			}

			mrInfo, err := s.getMrInfo(ctx, devFlow.AppID, currentBranch, targetBranch)
			if err != nil {
				return err
			}

			devFlowInfo.CodeNode = &pb.CodeNode{
				CurrentBranch:        currentBranch,
				Commit:               commitConvert(currentBranchCommit),
				IsJoinTempBranch:     devFlow.IsJoinTempBranch,
				JoinTempBranchStatus: devFlow.JoinTempBranchStatus,
				CanJoin:              canJoin(currentBranchCommit, baseCommit, tempBranch),
				Exist:                currentBranchExists,
				SourceBranch:         sourceBranch,
			}
			devFlowInfo.TempMergeNode = &pb.TempMergeNode{
				TempBranch:   tempBranch,
				BaseCommit:   commitConvert(baseCommit),
				ChangeBranch: appTempBranchChangeBranchListMap[fmt.Sprintf("%d%s", devFlow.AppID, tempBranch)],
			}
			devFlowInfo.PipelineNode = &pb.PipelineNode{PipelineStepInfos: pipelineStepInfo}

			devFlowInfo.MergeRequestNode = &pb.MergeRequestNode{
				CurrentBranch: currentBranch,
				TargetBranch:  targetBranch,
				Title:         s.makeMrTitle(ctx, currentBranch),
				Desc:          s.makeMrDesc(ctx, issue),
			}
			if mrInfo != nil {
				devFlowInfo.MergeRequestNode.Title = mrInfo.Title
				devFlowInfo.MergeRequestNode.Desc = mrInfo.Description
				devFlowInfo.MergeRequestNode.MergeRequestInfo = &gittarpb.MergeRequestInfo{
					Id:          mrInfo.Id,
					RepoMergeId: int64(mrInfo.RepoMergeId),
					Title:       mrInfo.Title,
					State:       mrInfo.State,
					AuthorId:    mrInfo.AuthorId,
					CloseAt:     pbutil.GetTimestamp(mrInfo.CloseAt),
					CloseUserId: mrInfo.CloseUserId,
					MergeAt:     pbutil.GetTimestamp(mrInfo.MergeAt),
					MergeUserId: mrInfo.MergeUserId,
				}
			}
			locker.Lock()
			devFlowInfos = append(devFlowInfos, &devFlowInfo)
			locker.Unlock()
			return nil
		}, index)
	}
	if work.Do().Error() != nil {
		return nil, work.Error()
	}

	sort.Slice(devFlowInfos, func(i, j int) bool {
		return devFlowInfos[i].DevFlow.CreatedAt.String() > devFlowInfos[j].DevFlow.CreatedAt.String()
	})

	return &pb.GetDevFlowInfoResponse{DevFlowInfos: devFlowInfos}, nil
}

func (s *Service) listDevFlowByReq(req *pb.GetDevFlowInfoRequest) (devFlows []db.DevFlow, err error) {
	if req.IssueID == 0 {
		return s.p.dbClient.ListDevFlowByAppIDAndBranch(req.AppID, req.Branch)
	}
	return s.p.dbClient.ListDevFlowByIssueID(req.IssueID)
}

func (s *Service) getMrInfo(ctx context.Context, appID uint64, currentBranch, targetBranch string) (mrInfo *apistructs.MergeRequestInfo, err error) {
	if currentBranch == "" || targetBranch == "" {
		return
	}
	result, err := s.p.bdl.ListMergeRequest(appID, apis.GetUserID(ctx), apistructs.GittarQueryMrRequest{
		TargetBranch: targetBranch,
		SourceBranch: currentBranch,
		Page:         1,
		Size:         999,
	})
	if err != nil {
		return nil, err
	}
	if len(result.List) > 0 {
		mrInfo = result.List[0]
	}
	return
}

func (s *Service) getAllAppInfoMapByProjectID(ctx context.Context, projectID uint64) (map[uint64]apistructs.ApplicationDTO, error) {
	var appInfoMap = map[uint64]apistructs.ApplicationDTO{}
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, err
	}
	appListResult, err := s.p.bdl.GetAppsByProjectSimple(projectID, uint64(orgID), apis.GetUserID(ctx))
	if err != nil {
		return nil, err
	}
	for _, app := range appListResult.List {
		appInfoMap[app.ID] = app
	}
	return appInfoMap, nil
}

func canJoin(commit, baseCommit *apistructs.Commit, tempBranch string) bool {
	if commit == nil {
		return false
	}
	if tempBranch == "" {
		return false
	}
	if baseCommit == nil {
		return true
	}
	if commit.ID == baseCommit.ID {
		return false
	}
	return true
}

func commitConvert(commit *apistructs.Commit) *pb.Commit {
	if commit == nil {
		return nil
	}
	var author, committer *pb.Signature
	if commit.Author != nil {
		author = &pb.Signature{
			Email: commit.Author.Email,
			Name:  commit.Author.Name,
			When:  timestamppb.New(commit.Author.When),
		}
	}
	if commit.Committer != nil {
		committer = &pb.Signature{
			Email: commit.Committer.Email,
			Name:  commit.Committer.Name,
			When:  timestamppb.New(commit.Committer.When),
		}
	}
	return &pb.Commit{
		ID:            commit.ID,
		Author:        author,
		Committer:     committer,
		CommitMessage: commit.CommitMessage,
		ParentSha:     commit.ParentSha,
	}
}

func (s *Service) queryAllPipelineYmlAndDoFunc(ctx context.Context, appDto *apistructs.ApplicationDTO, branch string,
	do func(ctx context.Context, locker *limit_sync_group.Locker, content string, node *apistructs.UnifiedFileTreeNode) error) error {

	ymlNodes, err := s.getAllPipelineYml(ctx, appDto, branch)
	if err != nil {
		return err
	}

	if len(ymlNodes) > 0 {
		ymlNodes = ymlNodes[:1]
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
	var pipelineNode []*pb.PipelineStepInfo

	err := s.queryAllPipelineYmlAndDoFunc(ctx, appDto, branch, func(ctx context.Context, locker *limit_sync_group.Locker, content string, node *apistructs.UnifiedFileTreeNode) error {
		pipelineYml, err := pipelineyml.New([]byte(content))
		if err != nil {
			s.p.Log.Errorf("failed to parse %v yaml:%v \n err:%v", content, pipelineYml, err)
			return nil
		}

		var stepInfo = pb.PipelineStepInfo{
			YmlName:         node.Name,
			HasOnPushBranch: pipelineYml.HasOnPushBranch(branch),
			PInode:          base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v/%v/tree/%v", appDto.ProjectID, appDto.ID, branch))),
		}
		stepInfo.Inode = node.Inode

		if !stepInfo.HasOnPushBranch {
			locker.Lock()
			pipelineNode = append(pipelineNode, &stepInfo)
			locker.Unlock()
			return nil
		}

		inodeBytes, err := base64.URLEncoding.DecodeString(node.Inode)
		if err != nil {
			return err
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
		pipelineNode = append(pipelineNode, &stepInfo)
		locker.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pipelineNode, nil
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
		if info.DevFlow.JoinTempBranchStatus == apistructs.JoinTempBranchFailedStatus {
			return "mergeFailed"
		}
		for _, pipelineInfo := range info.PipelineNode.PipelineStepInfos {
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
		s.p.Log.Errorf("failed to get %v path yml error %v", apistructs.DefaultPipelineYmlName, err)
		return nil, err
	}

	for _, node := range pipelineTreeData {
		if node.Name == apistructs.DefaultPipelineYmlName {
			nodes = append(nodes, node)
			break
		}
	}
	return nodes, nil
}
