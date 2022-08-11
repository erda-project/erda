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
	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/internal/apps/dop/dao"
)

type DBClient struct {
	*dao.DBClient
}

type RuleExecRecord struct {
	ID            string `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Scope         string
	ScopeID       string
	RuleID        string
	Code          string
	Env           Env
	Succeed       *bool
	ActionOutput  string
	SoftDeletedAt uint64
	Actor         string
}

func (RuleExecRecord) TableName() string {
	return "erda_rule_exec_history"
}

type Env map[string]interface{}

func (p Env) Value() (driver.Value, error) {
	if b, err := json.Marshal(p); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal env")
	} else {
		return string(b), nil
	}
}

func (p *Env) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for env")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, p); err != nil {
		return errors.Wrapf(err, "failed to unmarshal env")
	}
	return nil
}

func (db *DBClient) CreateRuleExecRecord(r *RuleExecRecord) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	r.ID = id.String()
	return db.Create(r).Error
}

func (db *DBClient) BatchCreateRuleExecRecords(r []RuleExecRecord) error {
	for i := range r {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		r[i].ID = id.String()
	}
	return db.BulkInsert(r)
}

func (db *DBClient) ListRuleExecRecords(req *pb.ListRuleExecHistoryRequest) ([]RuleExecRecord, int64, error) {
	var records []RuleExecRecord
	r := &RuleExecRecord{
		Scope:   req.Scope,
		ScopeID: req.ScopeID,
		RuleID:  req.RuleID,
		Succeed: req.Succeed,
	}
	var total int64
	offset := (req.PageNo - 1) * req.PageSize
	q := db.Model(&RuleExecRecord{}).Scopes(NotDeleted).Where(r).Order("created_at desc")
	if err := q.Offset(offset).Limit(req.PageSize).Find(&records).
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
