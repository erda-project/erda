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

package nexussvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/crypto/encryption"
)

// NexusSvc nexus 操作封装
type NexusSvc struct {
	db  *dao.DBClient
	bdl *bundle.Bundle

	rsaCrypt *encryption.RsaCrypt
}

type Option func(*NexusSvc)

func New(options ...Option) *NexusSvc {
	svc := &NexusSvc{}
	for _, op := range options {
		op(svc)
	}

	return svc
}

func WithDBClient(db *dao.DBClient) Option {
	return func(svc *NexusSvc) {
		svc.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *NexusSvc) {
		svc.bdl = bdl
	}
}

func WithRsaCrypt(rsaCrypt *encryption.RsaCrypt) Option {
	return func(svc *NexusSvc) {
		svc.rsaCrypt = rsaCrypt
	}
}
