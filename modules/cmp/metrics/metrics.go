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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/ierror"
	"github.com/erda-project/erda/pkg/i18n"
)

type ResourceType string

const (
	// SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name
	// usage rate , distribution rate , usage percent of distribution
	NodeCpuUsageSelectStatement    = `SELECT cpu_cores_usage::field FROM host_summary WHERE cluster_name::tag=$cluster_name AND host_ip::tag=$host_ip ORDER BY time DESC LIMIT 1`
	NodeMemoryUsageSelectStatement = `SELECT mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name AND host_ip::tag=$host_ip  ORDER BY time DESC LIMIT 1`
	PodCpuUsageSelectStatement     = `SELECT round_float(cpu_usage_percent::field, 2) FROM docker_container_summary WHERE pod_name::tag=$pod_name and pod_namespace::tag=$pod_namespace and podsandbox != true ORDER BY time DESC LIMIT 1`
	PodMemoryUsageSelectStatement  = `SELECT round_float(mem_usage_percent::field, 2) FROM docker_container_summary WHERE pod_name::tag=$pod_name and pod_namespace::tag=$pod_namespace and podsandbox != true ORDER BY time DESC LIMIT 1`

	Memory = "memory"
	Cpu    = "cpu"

	Pod  = "pod"
	Node = "node"
)

var (
	NilValueError     = errors.New("metrics nothing found")
	MetricsQueryError = errors.New("metrics nothing found")
)

type Metric struct {
	Metricq pb.MetricServiceServer
}

type MetricError struct {
	message  string
	code     string
	ctx      context.Context
	httpCode int
}

func (m MetricError) Render(locale *i18n.LocaleResource) string {
	return locale.Name()
}

func (m MetricError) Code() string {
	return m.code
}

func (m MetricError) HttpCode() int {
	return m.httpCode
}

func (m MetricError) Ctx() interface{} {
	return m.ctx
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

// QueryNodeResource query cpu and memory metrics from es database, return immediately if cache hit.
func (m *Metric) QueryNodeResource(ctx context.Context, req *apistructs.MetricsRequest) (httpserver.Responser, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		data []apistructs.MetricsData
		err  error
	)
	reqs := ToInfluxReq(req)
	for _, queryReq := range reqs {
		d := apistructs.MetricsData{}
		key := cache.GenerateKey([]string{queryReq.Params["host_ip"].String(), req.ClusterName, req.ResourceType})
		resp, err = m.DoQuery(ctx, key, queryReq)
		if err != nil {
			logrus.Errorf("internal error when query %v", queryReq)
		} else {
			if resp.Results[0].Series[0].Rows == nil {
				logrus.Errorf("internal error when query %v", queryReq)
			} else {
				d.Used = resp.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()
			}
		}
		data = append(data, d)
	}
	res := &apistructs.MetricsResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	}
	return mkResponse(res, nil), nil
}

func (m *Metric) QueryPodResource(ctx context.Context, req *apistructs.MetricsRequest) (httpserver.Responser, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		data []apistructs.MetricsData
		err  error
	)
	reqs := ToInfluxReq(req)
	for _, queryReq := range reqs {
		d := apistructs.MetricsData{}
		key := cache.GenerateKey([]string{queryReq.Params["pod_name"].String(), req.ResourceType})
		resp, err = m.DoQuery(ctx, key, queryReq)
		if err != nil {
			logrus.Errorf("internal error when query %v", queryReq)
		} else {
			if resp.Results[0].Series[0].Rows == nil {
				logrus.Errorf("internal error when query %v", queryReq)
			} else {
				d.Used = resp.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()
				d.Request = 0
				d.Total = 0
			}
		}
		data = append(data, d)
	}
	res := &apistructs.MetricsResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	}
	return mkResponse(res, nil), nil
}

func ToInfluxReq(req *apistructs.MetricsRequest) []*pb.QueryWithInfluxFormatRequest {
	queryReqs := make([]*pb.QueryWithInfluxFormatRequest, 0)
	//start, end, _ := getTimeRange("hour", 1, false)
	if req.ResourceKind == Node {
		for _, nreq := range req.NodeRequests {
			queryReq := &pb.QueryWithInfluxFormatRequest{}
			switch req.ResourceType {
			case Cpu:
				queryReq.Statement = NodeCpuUsageSelectStatement
			case Memory:
				queryReq.Statement = NodeMemoryUsageSelectStatement
			default:
				return nil
			}
			queryReq.Params = map[string]*structpb.Value{
				"cluster_name": structpb.NewStringValue(req.ClusterName),
				"host_ip":      structpb.NewStringValue(nreq.IP),
			}
			queryReqs = append(queryReqs, queryReq)
		}
	} else {
		for _, preq := range req.PodRequests {
			queryReq := &pb.QueryWithInfluxFormatRequest{}
			switch req.ResourceType {
			case Cpu:
				queryReq.Statement = PodCpuUsageSelectStatement
			case Memory:
				queryReq.Statement = PodMemoryUsageSelectStatement
			default:
				return nil
			}
			queryReq.Params = map[string]*structpb.Value{
				"pod_name":      structpb.NewStringValue(preq.PodName),
				"pod_namespace": structpb.NewStringValue(preq.Namespace),
			}
			queryReqs = append(queryReqs, queryReq)
		}
	}
	return queryReqs
}

func mkResponse(content interface{}, err ierror.IAPIError) httpserver.Responser {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: content,
		Error:   err,
	}
}
