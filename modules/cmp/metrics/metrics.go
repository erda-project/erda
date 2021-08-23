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
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cache"
)

type ResourceType string

const (
	// SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name
	CpuUsageSelectStatement = `SELECT cpu_cores_usage::field ,n_cpus::tag ,cpu_usage_active::field FROM host_summary WHERE cluster_name::tag=$cluster_name and hostname::tag=$hostname GROUP BY host_ip::tag`
	MemoryUsageSelectStatement = `SELECT mem_used::field ,mem_available::field,mem_free::field, mem_total::field  FROM host_summary  WHERE cluster_name::tag=$cluster_name and hostname::tag=$hostname GROUP BY host_ip::tag`
)

type Metric struct {
	Cache   *cache.Cache
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

//func queryExample(server pb.MetricServiceServer){
//	fmt.Println("query example")
//
//	req := &pb.QueryWithInfluxFormatRequest{
//		Start:     "before_1h", // or timestamp
//		End:       "now",       // or timestamp
//		Statement: `SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name GROUP BY host_ip::tag`,
//		Params: map[string]*structpb.Value{
//			"cluster_name": structpb.NewStringValue("terminus-dev"),
//		},
//	}
//	resp, err := server.QueryWithInfluxFormat(context.Background(), req)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(jsonx.MarshalAndIndent(resp))
//
//}
// curl cmp.project-387-dev:9027/api/metrics -X POST -H "Org-ID:1" -H "User-Id:2" -H "Content-type:application/json" -d "{\"cluster_name\":\"terminus-dev\",\"host_name\":[\"node-010000006216\"],\"resource_type\":\"cpu\"}"

// DoQuery query cpu and memory metrics from es database, return immediately if cache hit.
func (m *Metric) DoQuery(ctx context.Context, req apistructs.MetricsRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	var (
		expired  bool
		v        cache.Values
		err      error
		queryReq pb.QueryWithInfluxFormatRequest
		resp     *pb.QueryWithInfluxFormatResponse
		res      []*pb.QueryWithInfluxFormatResponse
	)

	switch req.ResourceType {
	case v1.ResourceCPU:
		queryReq.Statement = CpuUsageSelectStatement
	case v1.ResourceMemory:
		queryReq.Statement = MemoryUsageSelectStatement
	default:
		return nil, nil
	}

	queryReq.Start = "before_1h"
	queryReq.End =	"now"

	for _, name := range req.HostName {
		queryReq.Params = map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(req.ClusterName),
			"hostname":     structpb.NewStringValue(name),
		}

		key := cache.GenerateKey([]string{name, req.ClusterName, string(req.ResourceType)})
		if v, expired, err = m.Cache.Get(key); v != nil {
			logrus.Infof("%s hit cache, try return cache value directly",key)
			resp = &pb.QueryWithInfluxFormatResponse{}
			err = json.Unmarshal(v[0].(cache.ByteSliceValue).Value().([]byte), resp)
			if err != nil {
				logrus.Errorf("unmarshal failed")
			}
			if expired {
				logrus.Infof("cache expired")
				go func(ctx context.Context, key string, queryReq *pb.QueryWithInfluxFormatRequest, c *cache.Cache) {
					m.doQuery(ctx, key, m.Cache, queryReq)
				}(ctx, key, &queryReq, m.Cache)
			}
		} else {
			logrus.Infof("not hit cache, try fetch metrics")
			resp, err = m.doQuery(ctx, key, m.Cache, &queryReq)
			if err == nil {
				logrus.Error(err)
			}
		}
		res = append(res, resp)
	}
	return resp, nil
}
