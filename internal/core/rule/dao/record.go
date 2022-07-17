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

package dao

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/rule/pb"
	"github.com/erda-project/erda/internal/core/legacy/dao"
)

type DBClient struct {
	*dao.DBClient
}

type RuleSetExecRecord struct {
	ID            string `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Scope         string
	ScopeID       string
	RuleSetID     string
	Code          string
	Env           Env
	Succeed       bool
	ActionOutput  string
	SoftDeletedAt uint64
}

func (RuleSetExecRecord) TableName() string {
	return "erda_rule_set_exec_history"
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

func (db *DBClient) CreateRuleSetExecRecord(r *RuleSetExecRecord) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	r.ID = id.String()
	return db.Create(r).Error
}

func (db *DBClient) BatchCreateRuleSetExecRecords(r []RuleSetExecRecord) error {
	for i := range r {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		r[i].ID = id.String()
	}
	return db.BulkInsert(r)
}

func (db *DBClient) ListRuleSetExecRecords(req *pb.ListRuleSetExecHistoryRequest) ([]RuleSetExecRecord, int64, error) {
	var records []RuleSetExecRecord
	where := make(map[string]interface{})
	if req.Scope != "" {
		where["scope"] = req.Scope
	}
	if req.ScopeID != "" {
		where["scope_id"] = req.ScopeID
	}
	if req.EventType != "" {
		where["event_type"] = req.EventType
	}
	if req.Succeed != nil {
		where["succeed"] = req.Succeed
	}
	if req.RuleSetID != "" {
		where["rule_set_id"] = req.RuleSetID
	}
	var total int64
	offset := (req.PageNo - 1) * req.PageSize
	q := db.Model(&RuleSetExecRecord{}).Scopes(NotDeleted).Where(where).Order("created_at desc")
	if err := q.Offset(offset).Limit(req.PageSize).Find(&records).
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
