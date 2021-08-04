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

package log

import (
	"context"

	lru "github.com/hashicorp/golang-lru"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/pkg/common"
)

type Log struct {
	ctx   context.Context
	Log   pb.LogQueryServiceServer
	Cache *lru.Cache
}

func New() *Log {
	c := &Log{}
	c.Log = common.Hub.Service("erda.core.monitor.log.query.LogQueryService").(pb.LogQueryServiceServer)
	return c
}

func (c *Log) Query(ctx context.Context, req *pb.GetLogByRuntimeRequest) (*pb.GetLogByRuntimeResponse, error) {
	return c.Log.GetLogByRuntime(ctx, req)
}
