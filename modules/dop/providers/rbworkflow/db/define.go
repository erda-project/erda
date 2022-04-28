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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/plugin/soft_delete"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda-proto-go/dop/rbworkflow/pb"
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

type RBWorkflow struct {
	Model
	Scope
	Operator

	Stage       string
	Sort        uint64
	Branch      string
	Artifact    string
	Environment string
	SubFlows    SubFlows
}

type SubFlows []string

func (s SubFlows) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SubFlows) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &s)
}

func (RBWorkflow) TableName() string {
	return "erda_rb_workflow"
}

func (r *RBWorkflow) Convert() *pb.RbWorkflow {
	return &pb.RbWorkflow{
		ID:          r.ID.String,
		Stage:       r.Stage,
		Sort:        r.Sort,
		Branch:      r.Branch,
		Artifact:    r.Artifact,
		Environment: r.Environment,
		SubFlows:    r.SubFlows,
		OrgID:       r.OrgID,
		OrgName:     r.OrgName,
		ProjectID:   r.ProjectID,
		ProjectName: r.ProjectName,
		TimeCreated: timestamppb.New(r.CreatedAt),
		TimeUpdated: timestamppb.New(r.UpdatedAt),
		Creator:     r.Creator,
		Updater:     r.Updater,
	}
}

func (db Client) CreateWf(wf *RBWorkflow) error {
	return db.Create(wf).Error
}

func (db Client) MultiCreateWf(wfs []RBWorkflow) error {
	return db.Create(wfs).Error
}

func (db Client) GetWf(id string) (wf *RBWorkflow, err error) {
	err = db.Where("id = ?", id).First(&wf).Error
	return
}

func (db Client) DeleteWf(wf *RBWorkflow) error {
	return db.Delete(wf).Error
}

func (db Client) GetWfByProjectIDAndSort(proID, sort uint64) (wf *RBWorkflow, err error) {
	err = db.Model(&RBWorkflow{}).Where("project_id = ?", proID).Where("sort = ?", sort).First(&wf).Error
	return
}

func (db Client) GetWfByProjectIDAndStage(proID uint64, stage string) (wf *RBWorkflow, err error) {
	err = db.Model(&RBWorkflow{}).Where("project_id = ?", proID).Where("stage = ?", stage).First(&wf).Error
	return
}

func (db Client) ListWfByProjectID(proID uint64) (wfs []RBWorkflow, err error) {
	err = db.Model(&RBWorkflow{}).Where("project_id = ?", proID).Order("sort").Find(&wfs).Error
	return
}

func (db Client) UpdateWf(wf *RBWorkflow) error {
	return db.Save(wf).Error
}
