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

package gittar

import (
	"context"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/etcd"
)

type provider struct {
	ETCD       etcd.Interface   // autowired
	EtcdClient *clientv3.Client // autowired
}

func (p *provider) Run(ctx context.Context) error { return p.Initialize() }

func init() {
	servicehub.Register("gittar", &servicehub.Spec{
		Services:     []string{"gittar"},
		Dependencies: []string{"etcd"},
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
