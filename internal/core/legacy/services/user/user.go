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
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/legacy/dao"
)

type User struct {
	db *dao.DBClient
	uc userpb.UserServiceServer
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

func WithUCClient(uc userpb.UserServiceServer) Option {
	return func(o *User) {
		o.uc = uc
	}
}
