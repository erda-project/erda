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

package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/plugin/soft_delete"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
)

type Model struct {
	ID        fields.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt
}

type Scope struct {
	OrgID       uint64
	OrgName     string
	ProjectID   uint64
	ProjectName string
}

type Operator struct {
	Creator string
	Updater string
}

type DevFlowRule struct {
	Model
	Scope
	Operator

	Flows          JSON
	BranchPolicies JSON
}

func (DevFlowRule) TableName() string {
	return "erda_dev_flow_rule"
}

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) == 0 {
		return nil
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j JSON) String() string {
	return string(j)
}

type Flows []Flow

type Flow struct {
	Name         string `json:"name"`
	TargetBranch string `json:"targetBranch"`
	Artifact     string `json:"artifact"`
	Environment  string `json:"environment"`
}

func (f *Flow) Convert() *pb.Flow {
	return &pb.Flow{
		Name:         f.Name,
		TargetBranch: f.TargetBranch,
		Artifact:     f.Artifact,
		Environment:  f.Environment,
	}
}

type BranchPolicies []BranchPolicy

type (
	BranchPolicy struct {
		Branch     string  `json:"branch"`
		BranchType string  `json:"branchType"`
		Policy     *Policy `json:"policy"`
	}
	Policy struct {
		SourceBranch  string        `json:"sourceBranch"`
		CurrentBranch string        `json:"currentBranch"`
		TempBranch    string        `json:"tempBranch"`
		TargetBranch  *TargetBranch `json:"targetBranch"`
	}
	TargetBranch struct {
		MergeRequest string `json:"mergeRequest"`
		CherryPick   string `json:"cherryPick"`
	}
)

func (b *BranchPolicy) Convert() *pb.BranchPolicy {
	var (
		policy       *pb.Policy
		targetBranch *pb.TargetBranch
	)
	if b.Policy != nil && b.Policy.TargetBranch != nil {
		targetBranch = &pb.TargetBranch{
			MergeRequest: b.Policy.TargetBranch.MergeRequest,
			CherryPick:   b.Policy.TargetBranch.CherryPick,
		}
	}
	if b.Policy != nil {
		policy = &pb.Policy{
			SourceBranch:  b.Policy.SourceBranch,
			CurrentBranch: b.Policy.CurrentBranch,
			TempBranch:    b.Policy.TempBranch,
			TargetBranch:  targetBranch,
		}
	}

	return &pb.BranchPolicy{
		Branch:     b.Branch,
		BranchType: b.BranchType,
		Policy:     policy,
	}
}

func (r *DevFlowRule) Convert() (*pb.DevFlowRule, error) {
	flows := make([]*pb.Flow, 0)
	branchPolicies := make([]*pb.BranchPolicy, 0)
	if r.Flows != nil {
		if err := json.Unmarshal(r.Flows, &flows); err != nil {
			return nil, err
		}
	}

	if r.BranchPolicies != nil {
		if err := json.Unmarshal(r.BranchPolicies, &branchPolicies); err != nil {
			return nil, err
		}
	}

	return &pb.DevFlowRule{
		ID:             r.ID.String,
		OrgID:          r.OrgID,
		OrgName:        r.OrgName,
		ProjectID:      r.ProjectID,
		ProjectName:    r.ProjectName,
		TimeCreated:    timestamppb.New(r.CreatedAt),
		TimeUpdated:    timestamppb.New(r.UpdatedAt),
		Creator:        r.Creator,
		Updater:        r.Updater,
		Flows:          flows,
		BranchPolicies: branchPolicies,
	}, nil
}

func (db *Client) CreateDevFlowRule(f *DevFlowRule) error {
	return db.Create(f).Error
}

func (db *Client) GetDevFlowRule(id string) (f *DevFlowRule, err error) {
	err = db.Where("id = ?", id).First(&f).Error
	return
}

func (db *Client) GetDevFlowRuleByProjectID(proID uint64) (fs *DevFlowRule, err error) {
	err = db.Model(&DevFlowRule{}).Where("project_id = ?", proID).First(&fs).Error
	return
}

func (db *Client) UpdateDevFlowRule(f *DevFlowRule) error {
	return db.Save(f).Error
}

func (db *Client) DeleteDevFlowRuleByProjectID(projectID uint64) error {
	return db.Where("project_id = ?", projectID).Delete(&DevFlowRule{}).Error
}
