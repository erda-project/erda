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

package metrics

import (
	"context"

	lru "github.com/hashicorp/golang-lru"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/pkg/common"
)

type Metrics struct {
	ctx    context.Context
	Metric pb.MetricServiceServer
	Cache  *lru.Cache
}

func New() *Metrics {
	c := &Metrics{}
	c.Metric = common.Hub.Service("metric-query-example").(pb.MetricServiceServer)
	return c
}

func reqKey(req *pb.QueryWithInfluxFormatRequest, tag string) string {
	key := tag
	for _, v := range req.Params {
		key += v.String()
	}
	key = key + req.Start + req.End
	return key
}

func (c *Metrics) Query(req *pb.QueryWithInfluxFormatRequest, tag string) (*pb.QueryWithInfluxFormatResponse, error) {
	var (
		resp  *pb.QueryWithInfluxFormatResponse
		value interface{}
		ok    bool
		key   = reqKey(req, tag)
	)
	value, ok = c.Cache.Get(key)
	if ok {
		resp = value.(*pb.QueryWithInfluxFormatResponse)
		return resp, nil
	}
	format, err := c.Metric.QueryWithInfluxFormat(c.ctx, req)
	if err != nil {
		return nil, err
	}
	c.Cache.Add(key, format)
	return format, nil
}
