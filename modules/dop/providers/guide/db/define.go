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

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/guide/pb"
	"github.com/erda-project/erda/modules/dop/dao"
)

type GuideDB struct {
	*dao.DBClient
}

type GuideStatus string

const (
	InitStatus      GuideStatus = "init"
	ProcessedStatus GuideStatus = "processed"
	ExpiredStatus   GuideStatus = "expired"
)

const ExpiredTime = 6 * time.Hour

func (g GuideStatus) String() string {
	return string(g)
}

type GuideKind string

const (
	PipelineGuide GuideKind = "pipeline"
)

func (g GuideKind) String() string {
	return string(g)
}

type Guide struct {
	ID            string `gorm:"primary_key"`
	JumpLink      string
	Status        string
	Kind          string
	Creator       string
	OrgID         uint64
	OrgName       string
	ProjectID     uint64
	AppID         uint64
	Branch        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SoftDeletedAt uint64
}

func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("soft_deleted_at = 0")
}

func (Guide) TableName() string {
	return "erda_guide"
}

func (g *Guide) Convert() *pb.Guide {
	return &pb.Guide{
		ID:          g.ID,
		JumpLink:    g.JumpLink,
		Status:      g.Status,
		Creator:     g.Creator,
		Kind:        g.Kind,
		OrgID:       g.OrgID,
		OrgName:     g.OrgName,
		ProjectID:   g.ProjectID,
		AppID:       g.AppID,
		Branch:      g.Branch,
		TimeCreated: timestamppb.New(g.CreatedAt),
		TimeUpdated: timestamppb.New(g.UpdatedAt),
	}
}

// CreateGuide .
func (db *GuideDB) CreateGuide(guide *Guide) error {
	guide.ID = uuid.New().String()
	return db.Create(guide).Error
}

// GetGuide .
func (db *GuideDB) GetGuide(id string) (guide Guide, err error) {
	err = db.Debug().Model(&Guide{}).Scopes(NotDeleted).
		Where("id = ?", id).
		Where("status = ?", InitStatus).
		Where("created_at >= ?", time.Now().Add(-1*(ExpiredTime)).Format("2006-01-02 15:04:05")).
		First(&guide).Error
	return
}

// UpdateGuide .
func (db *GuideDB) UpdateGuide(id string, fields map[string]interface{}) error {
	guide := &Guide{ID: id}
	return db.Debug().Model(guide).Scopes(NotDeleted).Updates(fields).Error
}

// BatchUpdateGuideExpiryStatus .
func (db *GuideDB) BatchUpdateGuideExpiryStatus() error {
	return db.Debug().Model(&Guide{}).Scopes(NotDeleted).
		Where("status = ?", InitStatus).
		Where("created_at < ? ", time.Now().Add(-1*(ExpiredTime)).Format("2006-01-02 15:04:05")).
		Updates(map[string]interface{}{"status": ExpiredStatus}).Error
}

// UpdateGuideByAppIDAndBranch .
func (db *GuideDB) UpdateGuideByAppIDAndBranch(appID uint64, branch, kind string, fields map[string]interface{}) error {
	return db.Debug().Model(&Guide{}).Scopes(NotDeleted).
		Where("app_id = ?", appID).
		Where("branch = ?", branch).
		Where("kind = ?", kind).
		Updates(fields).Error
}

// ListGuide .
func (db *GuideDB) ListGuide(req *pb.ListGuideRequest, userID string) (guides []Guide, err error) {
	err = db.Debug().Model(&Guide{}).
		Scopes(NotDeleted).
		Where("kind = ?", req.Kind).
		Where("project_id = ?", req.ProjectID).
		Where("creator = ?", userID).
		Where("status = ?", InitStatus).
		Where("created_at >= ?", time.Now().Add(-1*(ExpiredTime)).Format("2006-01-02 15:04:05")).
		Order("created_at DESC").
		Limit(5).
		Find(&guides).Error
	return
}
