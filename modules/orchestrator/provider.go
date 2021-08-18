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

package orchestrator

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type provider struct {
	Election election.Interface `autowired:"etcd-election"`
	server   *httpserver.Server
	db       *dbclient.DBClient
}

func (p *provider) Init(ctx servicehub.Context) error {
	return p.Initialize()
}

func (p *provider) Run(ctx context.Context) error { return p.serve(ctx) }

func init() {
	servicehub.Register("orchestrator", &servicehub.Spec{
		Services:     []string{"orchestrator"},
		Dependencies: []string{"etcd-election"},
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
