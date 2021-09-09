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

package migration

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/pkg/ucauth"
)

type Migration struct {
	db *dao.DBClient
	uc *ucauth.UCClient
}

// Option 定义 Org 对象的配置选项
type Option func(*Migration)

// New 新建 Org 实例，通过 Org 实例操作企业资源
func New(options ...Option) *Migration {
	o := &Migration{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *Migration) {
		o.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(o *Migration) {
		o.uc = uc
	}
}

func (m *Migration) MigrateUser() error {
	users, err := m.db.GetUcUserList()
	if err != nil {
		return err
	}
	for _, u := range users {
		req := ucauth.OryKratosCreateIdentitiyRequest{
			SchemaID: "default",
			Traits: ucauth.OryKratosIdentityTraits{
				Email:  u.Email,
				Name:   u.Username,
				Nick:   u.Nickname,
				Phone:  u.Mobile,
				Avatar: u.Avatar,
			},
		}
		uuid, err := m.uc.UserMigration(req)
		if err != nil {
			logrus.Errorf("fail to migrate user: %v, err: %v", u.ID, err)
			continue
		}
		if err := m.db.InsertMapping(strconv.FormatInt(u.ID, 10), uuid, u.Password); err != nil {
			return err
		}
		logrus.Infof("migrate user %v to krataos user %v successfully", u.ID, uuid)
	}
	return nil
}
