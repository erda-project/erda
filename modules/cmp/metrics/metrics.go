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
	"fmt"
	"os"
	"strconv"
	"time"

	jsi "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/cmp/cache"
)

type ResourceType string

const (
	// SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name
	// usage rate , distribution rate , usage percent of distribution
	//NodeCpuUsageSelectStatement    = `SELECT last(cpu_cores_usage::field) FROM host_summary WHERE cluster_name::tag=$cluster_name AND host_ip::tag=$host_ip `
	NodeResourceUsageSelectStatement = `SELECT last(mem_used::field) as memRate , last(cpu_cores_usage::field) as cpuRate , host_ip::tag FROM host_summary WHERE cluster_name::tag=$cluster_name GROUP BY host_ip::tag`
	//NodeResourceUsageSelectStatement = `SELECT  mem_usage::field  ,cpu_cores_usage::field, host_ip FROM host_summary WHERE cluster_name::tag=$cluster_name GROUP BY host_ip::tag`
	//PodCpuUsageSelectStatement     = `SELECT SUM(cpu_allocation::field) * 100 / SUM(cpu_limit::field) as cpuRate, pod_name FROM docker_container_summary WHERE pod_namespace::tag=$pod_namespace and podsandbox != true GROUP BY pod_name::tag`
	PodResourceUsageSelectStatement = `SELECT round_float(SUM(mem_usage::field) * 100 / SUM(mem_limit::field),2) as memoryRate,round_float(SUM(cpu_allocation::field) * 100 / SUM(cpu_limit::field),2) as cpuRate, pod_name::tag, pod_namespace::tag FROM docker_container_summary WHERE cluster_name::tag=$cluster_name and podsandbox != true GROUP BY pod_name::tag`

	Memory = "memory"
	Cpu    = "cpu"

	Pod  = "pod"
	Node = "node"

	queryTimeout = 30 * time.Second
)

var (
	ResourceNotSupport = errors.New("resource type not support")
	QueryTimeoutError  = errors.New("metrics query timeout")
	ParameterNotFound  = errors.New("parameter required")
)

type Metric struct {
	ctx           context.Context
	Metricq       pb.MetricServiceServer
	metricReqChan chan []*MetricsReq
	limiter       *rate.Limiter
}

type MetricsReq struct {
	rawReq  *pb.QueryWithInfluxFormatRequest
	sync    bool
	resType string
	resKind string
	key     string
}

func New(metricq pb.MetricServiceServer, ctx context.Context) *Metric {
	m := &Metric{}
	m.ctx = ctx
	m.Metricq = metricq
	m.metricReqChan = make(chan []*MetricsReq, 1024)
	burst := 100
	if size, err := strconv.Atoi(os.Getenv("METRICS_QUEUE_SIZE")); err == nil && size > burst {
		burst = size
	}
	m.limiter = rate.NewLimiter(rate.Every(time.Millisecond), burst)
	go m.queryAsync()
	return m
}

func (m *Metric) queryAsync() {
	for {
		err := m.limiter.Wait(m.ctx)
		if err != nil {
			return
		}
		select {
		case reqs, able := <-m.metricReqChan:
			if !able {
				return
			} else if reqs == nil || len(reqs) == 0 {
				continue
			} else {
				metricReq := m.mergeReq(reqs)
				resp, err := m.query(m.ctx, metricReq.rawReq)
				if err != nil || resp == nil {
					logrus.Errorf("query metrics err,%v", err)
				} else {
					res := map[string]*MetricsData{}
					m.Store(resp, res, metricReq)
				}
			}
		}
	}
}

type Interface interface {
	NodeMetrics(ctx context.Context, req *MetricsRequest) (map[string]*MetricsData, error)
	PodMetrics(ctx context.Context, req *MetricsRequest) (map[string]*MetricsData, error)
}

func isEmptyResponse(resp *pb.QueryWithInfluxFormatResponse) bool {
	if resp == nil || len(resp.Results) == 0 ||
		len(resp.Results[0].Series) == 0 || len(resp.Results[0].Series[0].Rows) == 0 ||
		len(resp.Results[0].Series[0].Rows[0].Values) == 0 {
		return true
	}
	return false
}

func (m *Metric) query(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	return m.Metricq.QueryWithInfluxFormat(ctx, req)
}

// querySync returns influxdb data
func (m *Metric) querySync(ctx context.Context, req map[string][]*MetricsReq, c chan map[string]*MetricsData) {
	var (
		res  = make(map[string]*MetricsData)
		resp *pb.QueryWithInfluxFormatResponse
		err  error
	)

	for _, metricsReqs := range req {
		for _, mreq := range metricsReqs {
			d := &MetricsData{}
			res[mreq.key] = d
		}
		syncReqs := make([]*MetricsReq, 0)
		asyncReqs := make([]*MetricsReq, 0)
		for _, metricsReq := range metricsReqs {
			if metricsReq.sync {
				logrus.Infof("cache not found, try fetch metrics synchronized %v", metricsReq.key)
				syncReqs = append(syncReqs, metricsReq)
			} else {
				logrus.Infof("cache expired, try fetch metrics asynchronized %v", metricsReq.key)
				asyncReqs = append(asyncReqs, metricsReq)
			}
		}
		select {
		case m.metricReqChan <- asyncReqs:
		default:
			logrus.Warnf("channel is blocked, ayscn query skip")
		}
		metricsReq := m.mergeReq(syncReqs)
		err = m.limiter.Wait(ctx)
		if err != nil {
			logrus.Errorf("limiter error:%v", err)
			return
		}
		if metricsReq != nil {
			resp, err = m.query(ctx, metricsReq.rawReq)
			if err != nil || resp == nil {
				logrus.Errorf("metrics query error:%v, resp:%v", err, resp)
			} else {
				res = m.Store(resp, res, metricsReq)
			}
		}
	}
	c <- res
}

func (m *Metric) Store(resp *pb.QueryWithInfluxFormatResponse, res map[string]*MetricsData, metricsReq *MetricsReq) map[string]*MetricsData {
	if !isEmptyResponse(resp) {
		var (
			k    = ""
			d    *MetricsData
			tRes = make(map[string]*MetricsData)
		)
		for _, row := range resp.Results[0].Series[0].Rows {
			switch metricsReq.resKind {
			case Pod:
				k = cache.GenerateKey(Pod, Cpu, metricsReq.rawReq.Params["pod_namespace"].GetStringValue(), metricsReq.rawReq.Params["cluster_name"].GetStringValue(), row.Values[2].GetStringValue(), row.Values[3].GetStringValue())
				d = &MetricsData{
					Used: row.Values[1].GetNumberValue(),
				}
				tRes[k] = d
				if _, ok := res[k]; ok {
					res[k] = d
				}
				m.setCache(k, d)
				k = cache.GenerateKey(Pod, Memory, metricsReq.rawReq.Params["pod_namespace"].GetStringValue(), metricsReq.rawReq.Params["cluster_name"].GetStringValue(), row.Values[2].GetStringValue(), row.Values[3].GetStringValue())
				d = &MetricsData{
					Used: row.Values[0].GetNumberValue(),
				}
				tRes[k] = d
				if _, ok := res[k]; ok {
					res[k] = d
				}
				m.setCache(k, d)
			case Node:
				k = cache.GenerateKey(Node, Cpu, metricsReq.rawReq.Params["cluster_name"].GetStringValue(), row.Values[2].GetStringValue())
				d = &MetricsData{
					Used: row.Values[1].GetNumberValue(),
				}
				tRes[k] = d
				if _, ok := res[k]; ok {
					res[k] = d
				}
				m.setCache(k, d)

				k = cache.GenerateKey(Node, Memory, metricsReq.rawReq.Params["cluster_name"].GetStringValue(), row.Values[2].GetStringValue())
				d = &MetricsData{
					Used: row.Values[0].GetNumberValue(),
				}
				tRes[k] = d
				if _, ok := res[k]; ok {
					res[k] = d
				}
				m.setCache(k, d)
			}
		}
	}
	return res
}

func (m *Metric) setCache(k string, d interface{}) {
	data, err := cache.MarshalValue(d)
	if err != nil {
		logrus.Errorf("cache marshal metrics %v err: %v", k, err)
	} else {
		err := cache.GetFreeCache().Set(k, data, int64(time.Second*30))
		if err != nil {
			logrus.Errorf("cache set metrics %v err: %v", k, err)
		}
	}
}

func (m *Metric) mergeReq(reqs []*MetricsReq) *MetricsReq {
	if len(reqs) == 0 {
		return nil
	}
	var filterValues []interface{}

	for i := 0; i < len(reqs); i++ {
		filterValues = append(filterValues, reqs[i].rawReq.Filters[0].Value.GetListValue().Values[0].GetStringValue())
	}

	list, _ := structpb.NewList(filterValues)
	key := reqs[0].rawReq.Filters[0].Key
	op := reqs[0].rawReq.Filters[0].Op
	filter := &pb.Filter{}
	filter.Reset()
	filter.Key = key
	filter.Op = op
	filter.Value = structpb.NewListValue(list)
	reqs[0].rawReq.Filters = []*pb.Filter{
		filter,
	}

	return reqs[0]
}

// NodeMetrics query cpu and memory metrics from es database, return immediately if cache hit.
func (m *Metric) NodeMetrics(ctx context.Context, req *MetricsRequest) (map[string]*MetricsData, error) {
	var resp map[string]*MetricsData
	reqs, noNeed, err := ToInfluxReq(req, Node)
	if err != nil {
		return nil, err
	}
	ctx2, cancelFunc := context.WithTimeout(ctx, queryTimeout)
	defer cancelFunc()
	logrus.Infof("qeury node metrics with timeout")
	c := make(chan map[string]*MetricsData)
	go m.querySync(ctx2, reqs, c)
	select {
	case <-ctx2.Done():
		return nil, QueryTimeoutError
	case resp = <-c:
		for k, res := range resp {
			if res == nil {
				noNeed[k] = &MetricsData{}
			} else {
				noNeed[k] = res
			}
		}
	}
	return noNeed, nil
}

func (m *Metric) PodMetrics(ctx context.Context, req *MetricsRequest) (map[string]*MetricsData, error) {
	var resp map[string]*MetricsData
	reqs, noNeed, err := ToInfluxReq(req, Pod)
	if err != nil {
		return nil, err
	}
	ctx2, cancelFunc := context.WithTimeout(ctx, queryTimeout)
	defer cancelFunc()
	logrus.Infof("qeury pod metrics with timeout")
	c := make(chan map[string]*MetricsData)
	go m.querySync(ctx, reqs, c)
	select {
	case <-ctx2.Done():
		return nil, QueryTimeoutError
	case resp = <-c:
		for k, res := range resp {
			if res == nil {
				continue
			}
			noNeed[k] = res
		}
	}
	return noNeed, nil
}

func ToInfluxReq(req *MetricsRequest, kind string) (map[string][]*MetricsReq, map[string]*MetricsData, error) {
	queryReqs := make(map[string][]*MetricsReq)
	noNeed := make(map[string]*MetricsData)
	if req.Cluster == "" {
		return nil, nil, errors.New(fmt.Sprintf("parameter %s not found", req.Cluster))
	}
	if kind == Node {
		for _, nreq := range req.NodeRequests {
			key := nreq.CacheKey()
			if v, expired, err := cache.GetFreeCache().Get(key); err == nil {
				if v != nil {
					resp := &MetricsData{}
					err = jsi.Unmarshal(v[0].(cache.ByteSliceValue).Value().([]byte), resp)
					if err != nil {
						logrus.Errorf("try find cache error ,%v", err)
					}
					noNeed[key] = resp
					logrus.Infof("%v cache hit,isExpired %v", key, expired)
					if !expired {
						continue
					}
				}
				if v == nil || expired {
					queryReq := &pb.QueryWithInfluxFormatRequest{}
					queryReq.Start = "before_5m"
					queryReq.End = "now"
					queryReq.Statement = NodeResourceUsageSelectStatement
					queryReq.Params = map[string]*structpb.Value{
						"cluster_name": structpb.NewStringValue(nreq.ClusterName()),
						//"host_ip":      structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{structpb.NewStringValue(nreq.IP())}}),
					}
					queryReq.Filters = []*pb.Filter{{
						Key:   "tags.host_ip",
						Op:    "in",
						Value: structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{structpb.NewStringValue(nreq.IP())}}),
					}}
					sync := false
					if v == nil {
						sync = true
					}
					queryReqs[nreq.ClusterName()] = append(queryReqs[nreq.ClusterName()], &MetricsReq{rawReq: queryReq, sync: sync, resType: nreq.ResourceType(), resKind: nreq.ResourceKind(), key: nreq.CacheKey()})
				}
			} else {
				logrus.Errorf("Get %s cache error,%v", key, err)
			}
		}
	} else {
		for _, preq := range req.PodRequests {
			queryReq := &pb.QueryWithInfluxFormatRequest{}
			queryReq.Start = "before_5m"
			queryReq.End = "now"
			queryReq.Statement = PodResourceUsageSelectStatement
			queryReq.Params = map[string]*structpb.Value{
				//"pod_name":      structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{structpb.NewStringValue(preq.PodName())}}),
				"cluster_name": structpb.NewStringValue(preq.ClusterName()),
			}
			sync := false
			queryReqs[preq.Namespace()] = append(queryReqs[preq.Namespace()], &MetricsReq{rawReq: queryReq, sync: sync, resType: preq.ResourceType(), resKind: preq.ResourceKind(), key: preq.CacheKey()})
		}
	}
	return queryReqs, noNeed, nil
}
