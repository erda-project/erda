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

package apidocsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/apim/dbclient"
)

type Service struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

type Option func(service *Service)

func New(options ...Option) *Service {
	var service Service
	for _, op := range options {
		op(&service)
	}
	return &service
}

func WithDBClient(db *dbclient.DBClient) Option {
	return func(service *Service) {
		service.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(service *Service) {
		service.bdl = bdl
	}
}
