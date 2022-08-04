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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/plugin/soft_delete"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
)

type Model struct {
	ID        fields.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt
}

type Scope struct {
	OrgID   uint64
	OrgName string
	AppID   uint64
	AppName string
}

type Operator struct {
	Creator string
}

type DevFlow struct {
	Model
	Scope
	Operator

	Branch               string
	IssueID              uint64
	FlowRuleName         string
	JoinTempBranchStatus string
	IsJoinTempBranch     bool
}

func (DevFlow) TableName() string {
	return "erda_dev_flow"
}

func (f *DevFlow) Covert() *pb.DevFlow {
	return &pb.DevFlow{
		ID:                   f.ID.String,
		OrgID:                f.OrgID,
		OrgName:              f.OrgName,
		Creator:              f.Creator,
		Branch:               f.Branch,
		IssueID:              f.IssueID,
		FlowRuleName:         f.FlowRuleName,
		AppID:                f.AppID,
		AppName:              f.AppName,
		IsJoinTempBranch:     f.IsJoinTempBranch,
		JoinTempBranchStatus: f.JoinTempBranchStatus,
		CreatedAt:            timestamppb.New(f.CreatedAt),
		UpdatedAt:            timestamppb.New(f.UpdatedAt),
	}
}

func (db *Client) CreateDevFlow(f *DevFlow) error {
	return db.Create(f).Error
}

func (db *Client) GetDevFlow(id string) (f *DevFlow, err error) {
	err = db.Where("id = ?", id).First(&f).Error
	return
}

func (db *Client) ListDevFlowByIssueID(issueID uint64) (fs []DevFlow, err error) {
	err = db.Where("issue_id = ?", issueID).Find(&fs).Error
	return
}

func (db *Client) ListDevFlowByAppIDAndBranch(appID uint64, branch string) (fs []DevFlow, err error) {
	err = db.Where("app_id = ?", appID).Where("branch = ?", branch).Find(&fs).Error
	return
}

func (db *Client) ListDevFlowByFlowRuleName(flowRuleName string) (fs []DevFlow, err error) {
	err = db.Where("flow_rule_name = ?", flowRuleName).Find(&fs).Error
	return
}

func (db *Client) DeleteDevFlow(id string) error {
	return db.Where("id = ?", id).Delete(&DevFlow{}).Error
}

func (db *Client) UpdateDevFlow(f *DevFlow) error {
	return db.Save(f).Error
}
