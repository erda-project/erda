// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type ResourceType string

const (
	// SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name
	// usage rate , distribution rate , usage percent of distribution
	CpuUsageSelectStatement    = `SELECT cpu_cores_usage::field, cpu_request_total::field, n_cpus::tag FROM host_summary WHERE cluster_name::tag=$cluster_name and hostname::tag=$hostname GROUP BY host_ip::tag`
	MemoryUsageSelectStatement = `SELECT mem_used::field, mem_limit_total::field, mem_total::field  FROM host_summary  WHERE cluster_name::tag=$cluster_name and hostname::tag=$hostname GROUP BY host_ip::tag`
)

type MetricsServer interface {
	DoQuery(context.Context, *apistructs.MetricsRequest) ([]*apistructs.MetricsResponse, error)
}

type Metric struct {
	Metricq pb.MetricServiceServer
}

func (m *Metric) query(ctx context.Context, key string, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	v, err := m.Metricq.QueryWithInfluxFormat(ctx, req)

	if err != nil || v == nil {
		return nil, err
	}
	values, err := cache.MarshalValue(v)
	cache.FreeCache.Set(key, values, time.Now().UnixNano()+int64(time.Second*30))
	return v, nil
}

// DoQuery returns influxdb data
func (m *Metric) DoQuery(ctx context.Context, key string, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	var (
		expired bool
		v       cache.Values
		err     error
		resp    = &pb.QueryWithInfluxFormatResponse{}
	)

	if v, expired, err = cache.FreeCache.Get(key); v != nil {
		logrus.Infof("%s hit cache, try return cache value directly", key)
		err = json.Unmarshal(v[0].(cache.ByteSliceValue).Value().([]byte), resp)
		if err != nil {
			logrus.Errorf("unmarshal failed")
		}
		if expired {
			logrus.Infof("cache expired")
			go func(ctx context.Context, key string, queryReq *pb.QueryWithInfluxFormatRequest) {
				m.query(ctx, key, queryReq)
			}(ctx, key, req)
		}
	} else {
		logrus.Infof("not hit cache, try fetch metrics")
		resp, err = m.query(ctx, key, req)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
	}
	return resp, nil
}

// Query query cpu and memory metrics from es database, return immediately if cache hit.
func (m *Metric) Query(ctx context.Context, req *apistructs.MetricsRequest) (httpserver.Responser, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		data []apistructs.MetricsData
		err  error
	)
	reqs := ToInfluxReq(req)
	for _, queryReq := range reqs {
		d := apistructs.MetricsData{}
		key := cache.GenerateKey([]string{queryReq.Params["hostname"].String(), req.ClusterName, string(req.ResourceType)})
		resp, err = m.DoQuery(ctx, key, queryReq)
		if err != nil {
			// err occur, try next
			logrus.Error(err)
		} else {
			d.Used = resp.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()
			d.Request = resp.Results[0].Series[0].Rows[0].Values[1].GetNumberValue()
			d.Total = resp.Results[0].Series[0].Rows[0].Values[2].GetNumberValue()
		}
		data = append(data, d)
	}
	res := &apistructs.MetricsResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	}
	return mkResponse(res)
}

func ToInfluxReq(req *apistructs.MetricsRequest) []*pb.QueryWithInfluxFormatRequest {
	queryReqs := make([]*pb.QueryWithInfluxFormatRequest, 0)
	for _, name := range req.HostName {
		queryReq := &pb.QueryWithInfluxFormatRequest{}
		switch req.ResourceType {
		case v1.ResourceCPU:
			queryReq.Statement = CpuUsageSelectStatement
		case v1.ResourceMemory:
			queryReq.Statement = MemoryUsageSelectStatement
		default:
			return nil
		}

		queryReq.Start = "before_1m"
		queryReq.End = "now"
		queryReq.Params = map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(req.ClusterName),
			"hostname":     structpb.NewStringValue(name),
		}
		queryReqs = append(queryReqs, queryReq)
	}
	return queryReqs
}

func mkResponse(content interface{}) (httpserver.Responser, error) {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: content,
	}, nil
}
