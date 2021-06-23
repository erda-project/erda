// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package approval

import (
	"github.com/erda-project/erda/modules/core-services/dao"
)

type Approval struct {
	db *dao.DBClient
}

// Option .
type Option func(*Approval)

// New 新建 approval service
func New(options ...Option) *Approval {
	a := &Approval{}
	for _, op := range options {
		op(a)
	}
	return a
}

// WithDBClient 设置 dbclient
func WithDBClient(db *dao.DBClient) Option {
	return func(a *Approval) {
		a.db = db
	}
}
