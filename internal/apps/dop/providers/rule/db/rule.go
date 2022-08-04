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

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/dop/rule/pb"
)

type Rule struct {
	ID            string `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Scope         string
	ScopeID       string
	EventType     string
	Enabled       *bool
	Code          string
	Params        ActionParams
	Updator       string
	SoftDeletedAt uint64
}

type ActionParams pb.ActionParams

func (p ActionParams) Value() (driver.Value, error) {
	if b, err := json.Marshal(p); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal params")
	} else {
		return string(b), nil
	}
}

func (p *ActionParams) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for params")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, p); err != nil {
		return errors.Wrapf(err, "failed to unmarshal Extra")
	}
	return nil
}

func (Rule) TableName() string {
	return "erda_rule"
}

func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("soft_deleted_at = ?", 0)
}

func (db *DBClient) CreateRule(r *Rule) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	r.ID = id.String()
	return db.Create(r).Error
}

func (db *DBClient) GetRule(id string) (*Rule, error) {
	var r Rule
	err := db.Model(&Rule{}).Scopes(NotDeleted).Where("id = ?", id).First(&r).Error
	return &r, err
}

func (db *DBClient) DeleteRule(id string) error {
	return db.Model(&Rule{}).Scopes(NotDeleted).Where(&Rule{ID: id}).
		Update(map[string]interface{}{"soft_deleted_at": time.Now().UnixNano() / 1e6}).Error
}

func (db *DBClient) UpdateRule(r *Rule) error {
	return db.Model(&Rule{}).Scopes(NotDeleted).Where(&Rule{ID: r.ID}).Update(r).Error
}

func (db *DBClient) ListRules(req *pb.ListRulesRequest, withSystem bool) ([]Rule, int64, error) {
	var res []Rule
	r := &Rule{
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		EventType: req.EventType,
		Enabled:   req.Enabled,
		Updator:   req.Updator,
	}
	var total int64
	offset := (req.PageNo - 1) * req.PageSize
	q := db.Model(&Rule{}).Scopes(NotDeleted).Where(r).Order("updated_at desc")
	if withSystem {
		system := &Rule{
			Scope:     "system",
			EventType: req.EventType,
			Enabled:   req.Enabled,
		}
		q = q.Or(system)
	}
	if req.Name != "" {
		q = q.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if err := q.Offset(offset).Limit(req.PageSize).Find(&res).
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return res, total, nil
}
