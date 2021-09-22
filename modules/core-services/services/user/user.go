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

package user

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/pkg/ucauth"
)

type User struct {
	db *dao.DBClient
	uc *ucauth.UCClient
}

type Option func(*User)

func New(options ...Option) *User {
	o := &User{}
	for _, op := range options {
		op(o)
	}
	return o
}

func WithDBClient(db *dao.DBClient) Option {
	return func(o *User) {
		o.db = db
	}
}

func WithUCClient(uc *ucauth.UCClient) Option {
	return func(o *User) {
		o.uc = uc
	}
}

func (m *User) MigrateUser() error {
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

func (m *User) UcUserMigration() {
	if !conf.OryEnabled() {
		return
	}
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			if m.uc.MigrationReady() {
				if err := m.MigrateUser(); err != nil {
					logrus.Errorf("fail to migrate user, %v", err)
				}
				return
			}
		}
	}
}
