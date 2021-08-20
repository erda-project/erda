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
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cache"
)

type ResourceType string

const (
	CpuUsageSelectStatement = `SELECT cpu_cores_usage ,n_cpus ,cpu_usage_active FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`
	MemoryUsageSelectStatement = `SELECT mem_used ,mem_available,mem_free, mem_total  FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`
)

type Metric struct {
	cache   *cache.Cache
	Metricq pb.MetricServiceServer
}

func (m *Metric) doQuery(ctx context.Context, key string, c *cache.Cache, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	v, err := m.Metricq.QueryWithInfluxFormat(ctx, req)

	if err != nil {
		return nil, err
	}
	values, err := cache.MarshalValue(v)
	if err != nil {
		if err := c.Set(key, values, time.Now().UnixNano()+int64(time.Second*30)); err != nil {
			return nil, err
		}
	}
	return v, nil

}

func (m *Metric) DoQuery(ctx context.Context, req apistructs.MetricsRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	var (
		expired  bool
		v        cache.Values
		err      error
		queryReq pb.QueryWithInfluxFormatRequest
		resp     *pb.QueryWithInfluxFormatResponse
		res     []*pb.QueryWithInfluxFormatResponse
	)
	start := time.Now().UnixNano()
	switch req.ResourceType {
	case v1.ResourceCPU:
		queryReq.Statement = CpuUsageSelectStatement
	case v1.ResourceMemory:
		queryReq.Statement = MemoryUsageSelectStatement
	default:
		return nil, nil
	}
	queryReq.Start = fmt.Sprintf("%d", start-int64(30*time.Second))
	queryReq.End = fmt.Sprintf("%d", start)
	for _, name := range req.HostName {

		queryReq.Params = map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(req.ClusterName),
			"hostname":     structpb.NewStringValue(name),
		}

		key := cache.GenerateKey([]string{name, req.ClusterName, string(req.ResourceType)})

		if v, expired, err = m.cache.Get(key); v != nil {
			resp = &pb.QueryWithInfluxFormatResponse{}
			err := json.Unmarshal(v[0].(cache.ByteSliceValue).Value().([]byte), &resp)
			if err != nil {
				return nil, err
			}
			if expired {
				go func(ctx context.Context, key string, queryReq *pb.QueryWithInfluxFormatRequest, c *cache.Cache) {
					m.doQuery(ctx, key, m.cache, queryReq)
				}(ctx, key, &queryReq, m.cache)
			}
		} else {
			resp, err = m.doQuery(ctx, key, m.cache, &queryReq)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, resp)
	}

	return resp, nil
}
