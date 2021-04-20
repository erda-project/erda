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

package cq

import "github.com/erda-project/erda/bundle"

type CQ struct {
	bdl *bundle.Bundle
}

type Option func(*CQ)

func New(options ...Option) *CQ {
	var cq CQ
	for _, op := range options {
		op(&cq)
	}
	return &cq
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(cq *CQ) {
		cq.bdl = bdl
	}
}
