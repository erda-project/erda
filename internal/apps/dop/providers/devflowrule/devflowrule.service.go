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

package devflowrule

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/devflowrule/db"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

const (
	resource = "devFlowRule"
)

type BranchType string

const (
	MultiBranchType  BranchType = "multi_branch"
	SingleBranchType BranchType = "single_branch"
)

func (b BranchType) String() string {
	return string(b)
}

type GetFlowByRuleRequest struct {
	ProjectID     uint64
	BranchType    string
	CurrentBranch string
	SourceBranch  string
}

type Interface interface {
	pb.DevFlowRuleServiceServer
	GetFlowByRule(context.Context, GetFlowByRuleRequest) (*pb.FlowWithBranchPolicy, error)
}

func (p *provider) CreateDevFlowRule(ctx context.Context, request *pb.CreateDevFlowRuleRequest) (*pb.CreateDevFlowRuleResponse, error) {
	project, err := p.bundle.GetProject(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	orgResp, err := p.Org.GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcDOP),
		&orgpb.GetOrgRequest{IdOrName: strconv.FormatUint(project.OrgID, 10)})
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	org := orgResp.Data

	flows := p.InitFlows()
	b, err := json.Marshal(&flows)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}

	policies := p.InitBranchPolicies()
	policiesByte, err := json.Marshal(&policies)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}

	devFlow := db.DevFlowRule{
		Scope: db.Scope{
			OrgID:       org.ID,
			OrgName:     org.Name,
			ProjectID:   project.ID,
			ProjectName: project.Name,
		},
		Operator: db.Operator{
			Creator: request.UserID,
			Updater: request.UserID,
		},
		Flows:          b,
		BranchPolicies: policiesByte,
	}
	err = p.dbClient.CreateDevFlowRule(&devFlow)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	data, err := devFlow.Convert()
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	return &pb.CreateDevFlowRuleResponse{Data: data}, nil
}

func (p *provider) InitFlows() db.Flows {
	return db.Flows{
		{
			Name:         "DEV",
			TargetBranch: "feature/*",
			Artifact:     "alpha",
			Environment:  "DEV",
		},
		{
			Name:         "TEST",
			TargetBranch: "develop",
			Artifact:     "beta",
			Environment:  "TEST",
		},
		{
			Name:         "STAGING",
			TargetBranch: "release/*",
			Artifact:     "rc",
			Environment:  "STAGING",
		},
		{
			Name:         "PROD",
			TargetBranch: "master",
			Artifact:     "stable",
			Environment:  "PROD",
		},
	}
}

func (p *provider) InitBranchPolicies() db.BranchPolicies {
	return db.BranchPolicies{
		{
			Branch:     "feature/*",
			BranchType: SingleBranchType.String(),
			Policy:     nil,
		},
		{
			Branch:     "develop",
			BranchType: SingleBranchType.String(),
			Policy:     nil,
		},
		{
			Branch:     "release/*",
			BranchType: SingleBranchType.String(),
			Policy:     nil,
		},
		{
			Branch:     "master",
			BranchType: SingleBranchType.String(),
			Policy:     nil,
		},
	}
}

func (p *provider) UpdateDevFlowRule(ctx context.Context, request *pb.UpdateDevFlowRuleRequest) (*pb.UpdateDevFlowRuleResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InvalidParameter(err)
	}

	devFlow, err := p.dbClient.GetDevFlowRule(request.ID)
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	access, err := p.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  devFlow.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrUpdateDevFlowRule.AccessDenied()
	}

	if err = p.CheckFlow(request.Flows); err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}
	flowsByte, err := json.Marshal(request.Flows)
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	policesByte, err := json.Marshal(request.BranchPolicies)
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	devFlow.Flows = flowsByte
	devFlow.BranchPolicies = policesByte
	devFlow.Operator.Updater = apis.GetUserID(ctx)
	if err = p.dbClient.UpdateDevFlowRule(devFlow); err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	data, err := devFlow.Convert()
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}
	return &pb.UpdateDevFlowRuleResponse{Data: data}, nil
}

func (p *provider) CheckFlow(flows []*pb.Flow) error {
	nameMap := make(map[string]struct{})
	for _, v := range flows {
		if _, ok := nameMap[v.Name]; ok {
			return fmt.Errorf("the name %s is duplicate", v.Name)
		} else {
			nameMap[v.Name] = struct{}{}
		}
	}
	return nil
}

func (p *provider) GetDevFlowRulesByProjectID(ctx context.Context, request *pb.GetDevFlowRuleRequest) (*pb.GetDevFlowRuleResponse, error) {
	if !apis.IsInternalClient(ctx) {
		access, err := p.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   apis.GetUserID(ctx),
			Scope:    apistructs.ProjectScope,
			ScopeID:  request.ProjectID,
			Resource: resource,
			Action:   apistructs.ListAction,
		})
		if err != nil {
			return nil, apierrors.ErrGetDevFlowRule.InternalError(err)
		}
		if !access.Access {
			return nil, apierrors.ErrGetDevFlowRule.AccessDenied()
		}
	}

	wfs, err := p.dbClient.GetDevFlowRuleByProjectID(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrGetDevFlowRule.InternalError(err)
	}
	data, err := wfs.Convert()
	if err != nil {
		return nil, apierrors.ErrGetDevFlowRule.InternalError(err)
	}
	return &pb.GetDevFlowRuleResponse{Data: data}, nil
}

func (p *provider) DeleteDevFlowRule(ctx context.Context, request *pb.DeleteDevFlowRuleRequest) (*pb.DeleteDevFlowRuleResponse, error) {
	if err := p.dbClient.DeleteDevFlowRuleByProjectID(request.ProjectID); err != nil {
		return nil, apierrors.ErrDeleteDevFlowRule.InternalError(err)
	}
	return &pb.DeleteDevFlowRuleResponse{}, nil
}

func (p *provider) GetFlowByRule(ctx context.Context, request GetFlowByRuleRequest) (*pb.FlowWithBranchPolicy, error) {
	wfs, err := p.dbClient.GetDevFlowRuleByProjectID(request.ProjectID)
	if err != nil {
		return nil, err
	}
	flows := make(db.Flows, 0)
	if err = json.Unmarshal(wfs.Flows, &flows); err != nil {
		return nil, err
	}
	branchPolicies := make(db.BranchPolicies, 0)
	if err = json.Unmarshal(wfs.BranchPolicies, &branchPolicies); err != nil {
		return nil, err
	}

	var (
		findBranch string
		findFlow   db.Flow
	)
	for _, flow := range flows {
		targetBranches := strings.Split(flow.TargetBranch, ",")
		for _, branch := range targetBranches {
			if diceworkspace.IsRefPatternMatch(request.CurrentBranch, []string{branch}) {
				findBranch = flow.TargetBranch
				findFlow = flow
				break
			}
		}
	}
	if findBranch == "" {
		return nil, nil
	}
	for _, policy := range branchPolicies {
		if findBranch == policy.Branch && request.BranchType == policy.BranchType && policy.Policy != nil {
			sourceBranches := strings.Split(policy.Policy.SourceBranch, ",")
			if diceworkspace.IsRefPatternMatch(request.SourceBranch, sourceBranches) {
				return &pb.FlowWithBranchPolicy{
					Flow:         findFlow.Convert(),
					BranchPolicy: policy.Convert(),
				}, nil
			}
		}
	}
	return nil, nil
}
