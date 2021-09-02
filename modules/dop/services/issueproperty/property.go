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
