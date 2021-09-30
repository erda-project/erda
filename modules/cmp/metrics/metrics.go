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
	"net/http"
	"os"
	"strconv"
	"time"

	jsi "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/ierror"
	"github.com/erda-project/erda/pkg/i18n"
)

type ResourceType string

var queryQueue chan struct{}

func init() {
	queueSize := 50
	if size, err := strconv.Atoi(os.Getenv("METRICS_QUEUE_SIZE")); err == nil && size > queueSize {
		queueSize = size
	}
	queryQueue = make(chan struct{}, queueSize)
}

const (
	// SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name
	// usage rate , distribution rate , usage percent of distribution
	NodeCpuUsageSelectStatement    = `SELECT last(cpu_cores_usage::field) FROM host_summary WHERE cluster_name::tag=$cluster_name AND host_ip::tag=$host_ip and time > now() -300s`
	NodeMemoryUsageSelectStatement = `SELECT last(mem_used::field) FROM host_summary WHERE cluster_name::tag=$cluster_name AND host_ip::tag=$host_ip  and time > now() -300s`
	PodCpuUsageSelectStatement     = `SELECT round_float(last(cpu_usage_percent::field), 2) FROM docker_container_summary WHERE pod_name::tag=$pod_name and pod_namespace::tag=$pod_namespace and podsandbox != true and time > now() -300s`
	PodMemoryUsageSelectStatement  = `SELECT round_float(last(mem_usage_percent::field), 2) FROM docker_container_summary WHERE pod_name::tag=$pod_name and pod_namespace::tag=$pod_namespace and podsandbox != true and time > now() -300s`

	Memory = "memory"
	Cpu    = "cpu"

	Pod  = "pod"
	Node = "node"
)

var (
	ResourceNotSupport = errors.New("resource type not support")
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

type Interface interface {
	NodeMetrics(ctx context.Context, req *MetricsRequest) ([]MetricsData, error)
	PodMetrics(ctx context.Context, req *MetricsRequest) ([]MetricsData, error)
}

func (m *Metric) query(ctx context.Context, key string, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	queryQueue <- struct{}{}
	v, err := m.Metricq.QueryWithInfluxFormat(ctx, req)
	<-queryQueue
	if err != nil {
		return nil, err
	}
	values, err := cache.MarshalValue(v)
	cache.FreeCache.Set(key, values, time.Now().UnixNano()+int64(time.Second*30))
	return v, nil
}

// DoQuery returns influxdb data
func (m *Metric) DoQuery(ctx context.Context, cacheKey string, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	var (
		expired bool
		v       cache.Values
		err     error
		resp    = &pb.QueryWithInfluxFormatResponse{}
	)
	logrus.Infof("start get metrics of %s", cacheKey)
	if v, expired, err = cache.FreeCache.Get(cacheKey); v != nil {
		logrus.Infof("%s hit cache, return cache value directly", cacheKey)
		err = jsi.Unmarshal(v[0].(cache.ByteSliceValue).Value().([]byte), resp)
		if err != nil {
			logrus.Errorf("unmarshal failed")
		}
		if expired {
			logrus.Infof("cache expired, try fetch metrics asynchronized")
			go func(ctx context.Context, key string, queryReq *pb.QueryWithInfluxFormatRequest) {
				m.query(ctx, key, queryReq)
			}(ctx, cacheKey, req)
		}
	} else {
		logrus.Infof("%s not hit cache, try fetch metrics synchronized", cacheKey)
		resp, err = m.query(ctx, cacheKey, req)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
	}
	return resp, nil
}

// NodeMetrics query cpu and memory metrics from es database, return immediately if cache hit.
func (m *Metric) NodeMetrics(ctx context.Context, req *MetricsRequest) ([]MetricsData, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		data []MetricsData
		err  error
	)
	reqs, err := ToInfluxReq(req)
	if err != nil {
		return nil, err
	}
	for _, queryReq := range reqs {
		d := MetricsData{}
		key := cache.GenerateKey([]string{queryReq.Params["host_ip"].String(), req.ClusterName, req.ResourceType})
		resp, err = m.DoQuery(ctx, key, queryReq)
		if err != nil {
			logrus.Errorf("internal error when query %v", queryReq)
		} else {
			if resp.Results[0].Series[0].Rows == nil {
				logrus.Warnf("result empty when query %v", queryReq)
			} else {
				d.Used = resp.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()
			}
		}
		data = append(data, d)
	}
	return data, nil
}

func (m *Metric) PodMetrics(ctx context.Context, req *MetricsRequest) ([]MetricsData, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		data []MetricsData
		err  error
	)
	reqs, err := ToInfluxReq(req)
	if err != nil {
		return nil, err
	}
	for _, queryReq := range reqs {
		d := MetricsData{}
		key := cache.GenerateKey([]string{queryReq.Params["pod_name"].String(), req.ResourceType})
		resp, err = m.DoQuery(ctx, key, queryReq)
		if err != nil {
			logrus.Errorf("internal error when query %v", queryReq)
		} else {
			if resp.Results[0].Series[0].Rows == nil {
				logrus.Errorf("result empty when query %v", queryReq)
			} else {
				d.Used = resp.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()
				d.Request = 0
				d.Total = 0
			}
		}
		data = append(data, d)
	}
	return data, nil
}

func ToInfluxReq(req *MetricsRequest) ([]*pb.QueryWithInfluxFormatRequest, error) {
	queryReqs := make([]*pb.QueryWithInfluxFormatRequest, 0)
	if req.ResourceKind == Node {
		for _, nreq := range req.NodeRequests {
			queryReq := &pb.QueryWithInfluxFormatRequest{}
			switch req.ResourceType {
			case Cpu:
				queryReq.Statement = NodeCpuUsageSelectStatement
			case Memory:
				queryReq.Statement = NodeMemoryUsageSelectStatement
			default:
				return nil, ResourceNotSupport
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
				return nil, ResourceNotSupport
			}
			queryReq.Params = map[string]*structpb.Value{
				"pod_name":      structpb.NewStringValue(preq.PodName),
				"pod_namespace": structpb.NewStringValue(preq.Namespace),
			}
			queryReqs = append(queryReqs, queryReq)
		}
	}
	return queryReqs, nil
}

func mkResponse(content interface{}, err ierror.IAPIError) httpserver.Responser {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: content,
		Error:   err,
	}
}
