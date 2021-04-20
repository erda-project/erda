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

package reportsvc

import (
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

type ReportSvc struct {
	dbClient *dbclient.Client
}

func New(ops ...Option) *ReportSvc {
	var svc ReportSvc

	for _, op := range ops {
		op(&svc)
	}

	return &svc
}

type Option func(*ReportSvc)

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(svc *ReportSvc) {
		svc.dbClient = dbClient
	}
}
