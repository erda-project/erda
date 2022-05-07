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
	"github.com/erda-project/erda-proto-go/dop/devworkflow/pb"
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

type DevWorkflow struct {
	Model
	Scope
	Operator

	WorkFlows JSON
}

func (DevWorkflow) TableName() string {
	return "erda_dev_workflow"
}

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
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

type WorkFlows []WorkFlow

type WorkFlow struct {
	Name               string              `json:"name"`
	FlowType           string              `json:"flowType"`
	TargetBranch       string              `json:"targetBranch"`
	ChangeFromBranch   string              `json:"changeFromBranch"`
	ChangeBranch       string              `json:"changeBranch"`
	EnableAutoMerge    bool                `json:"enableAutoMerge"`
	AutoMergeBranch    string              `json:"autoMergeBranch"`
	Artifact           string              `json:"artifact"`
	Environment        string              `json:"environment"`
	StartWorkflowHints []StartWorkflowHint `json:"startWorkflowHints"`
}

type StartWorkflowHint struct {
	Place            string
	ChangeBranchRule string
}

func (r *DevWorkflow) Convert() *pb.DevWorkflow {
	return &pb.DevWorkflow{
		ID:          r.ID.String,
		OrgID:       r.OrgID,
		OrgName:     r.OrgName,
		ProjectID:   r.ProjectID,
		ProjectName: r.ProjectName,
		TimeCreated: timestamppb.New(r.CreatedAt),
		TimeUpdated: timestamppb.New(r.UpdatedAt),
		Creator:     r.Creator,
		Updater:     r.Updater,
		WorkFlows:   r.WorkFlows.String(),
	}
}

func (db Client) CreateWf(wf *DevWorkflow) error {
	return db.Create(wf).Error
}

func (db Client) GetWf(id string) (wf *DevWorkflow, err error) {
	err = db.Where("id = ?", id).First(&wf).Error
	return
}

func (db Client) GetWfByProjectID(proID uint64) (wfs *DevWorkflow, err error) {
	err = db.Model(&DevWorkflow{}).Where("project_id = ?", proID).First(&wfs).Error
	return
}

func (db Client) UpdateWf(wf *DevWorkflow) error {
	return db.Save(wf).Error
}

func (db Client) DeleteWfByProjectID(projectID uint64) error {
	return db.Where("project_id = ?", projectID).Delete(&DevWorkflow{}).Error
}
