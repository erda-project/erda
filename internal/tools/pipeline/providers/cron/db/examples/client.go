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

package main

import (
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type PipelineCron struct {
	ID            uint64    `json:"id" xorm:"pk autoincr"`
	TimeCreated   time.Time `json:"timeCreated" xorm:"created"`
	TimeUpdated   time.Time `json:"timeUpdated" xorm:"updated"`
	SoftDeletedAt int64     `json:"softDeletedAt" xorm:"deleted notnull"`
}

func (p PipelineCron) TableName() string {
	return "pipeline_crons"
}

type Client struct {
	mysqlxorm.Interface
}

func (client *Client) GetPipelineCron(id uint64, ops ...mysqlxorm.SessionOption) (cron PipelineCron, bool bool, err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	defer func() {
		err = errors.Wrapf(err, "failed to get pipeline cron by id [%v]", id)
	}()

	found, err := session.ID(id).Get(&cron)
	if err != nil {
		return PipelineCron{}, false, err
	}
	if !found {
		return PipelineCron{}, false, nil
	}
	return cron, true, nil
}

func (client *Client) CreatePipelineCron(cron *PipelineCron, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if cron.ID == 0 {
		cron.ID = uuid.SnowFlakeIDUint64()
	}
	_, err := session.Insert(cron)
	return errors.Wrapf(err, "failed to create pipeline cron")
}

func (client *Client) DeletePipelineCron(id uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if _, err := session.ID(id).Delete(&PipelineCron{}); err != nil {
		return errors.Errorf("failed to delete pipeline cron, id: %d, err: %v", id, err)
	}
	return nil
}

func (client *Client) BatchDeletePipelineCron(ids []uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if len(ids) <= 0 {
		return nil
	}

	if _, err := session.In("id", ids).Delete(&PipelineCron{}); err != nil {
		return errors.Errorf("failed to delete pipeline cron, ids: %d, err: %v", ids, err)
	}
	return nil
}
