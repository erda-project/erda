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

// Package asset API 资产
package assetsvc

import "github.com/erda-project/erda/modules/dop/services/branchrule"

type Service struct {
	branchRuleSvc *branchrule.BranchRule
}

type Option func(*Service)

func New(options ...Option) *Service {
	r := &Service{}
	for _, op := range options {
		op(r)
	}
	return r
}

func WithBranchRuleSvc(svc *branchrule.BranchRule) Option {
	return func(service *Service) {
		service.branchRuleSvc = svc
	}
}
