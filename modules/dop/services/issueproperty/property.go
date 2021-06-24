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

package issueproperty

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
)

// IssueProperty issue property 对象
type IssueProperty struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 IssueProperty 对象配置选项
type Option func(*IssueProperty)

// New 新建 issue stream 对象
func New(options ...Option) *IssueProperty {
	is := &IssueProperty{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssueProperty) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssueProperty) {
		is.bdl = bdl
	}
}
