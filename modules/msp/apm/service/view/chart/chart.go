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

package chart

import (
	"context"
	"fmt"
	"strings"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/service/pb"
)

const Layout = "2006-01-02T15:04:05Z"

type Chart interface {
	GetChart(ctx context.Context) (*pb.ServiceChart, error)
}

type TransactionLayerType string

const (
	TransactionLayerHttp  TransactionLayerType = "http"
	TransactionLayerRpc   TransactionLayerType = "rpc"
	TransactionLayerCache TransactionLayerType = "cache"
	TransactionLayerDb    TransactionLayerType = "db"
	TransactionLayerMq    TransactionLayerType = "mq"
)

type BaseChart struct {
	StartTime int64
	EndTime   int64
	Interval  string
	TenantId  string
	ServiceId string
	Layers    []TransactionLayerType
	LayerPath string
	Metric    metricpb.MetricServiceServer
}

func (b *BaseChart) getDataSourceNames() string {
	var list []string
	for _, layer := range b.Layers {
		switch layer {
		case TransactionLayerHttp:
			list = append(list, "application_http")
		case TransactionLayerRpc:
			list = append(list, "application_rpc")
		case TransactionLayerCache:
			list = append(list, "application_cache")
		case TransactionLayerDb:
			list = append(list, "application_db")
		case TransactionLayerMq:
			list = append(list, "application_mq")
		}
	}
	return strings.Join(list, ",")
}

func (b *BaseChart) getLayerPathKeys() []string {
	var list []string
	for _, layer := range b.Layers {
		switch layer {
		case TransactionLayerHttp:
			list = append(list, "http_path::tag")
		case TransactionLayerRpc:
			list = append(list, "rpc_method::tag")
		case TransactionLayerCache:
			list = append(list, "db_statement::tag")
		case TransactionLayerDb:
			list = append(list, "db_statement::tag")
		case TransactionLayerMq:
			list = append(list, "message_bus_destination::tag")
		}
	}
	return list
}

func (b *BaseChart) buildLayerPathFilterSql(paramName string) string {
	if len(b.LayerPath) == 0 {
		return ""
	}

	keys := b.getLayerPathKeys()
	var tokens []string
	for _, key := range keys {
		tokens = append(tokens, fmt.Sprintf("%s=%s", key, paramName))
	}

	return fmt.Sprintf("AND (%s) ", strings.Join(tokens, " OR "))
}

func Selector(chartType string, baseChart *BaseChart, ctx context.Context) (*pb.ServiceChart, error) {
	switch chartType {
	case strings.ToLower(pb.ChartType_HttpCode.String()):
		httpCodeChart := HttpCodeChart{BaseChart: baseChart}
		getChart, err := httpCodeChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	case strings.ToLower(pb.ChartType_RPS.String()):
		rpcChart := RpsChart{BaseChart: baseChart}
		getChart, err := rpcChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	case strings.ToLower(pb.ChartType_AvgDuration.String()):
		avgDurationChart := AvgDurationChart{BaseChart: baseChart}
		getChart, err := avgDurationChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	case strings.ToLower(pb.ChartType_ErrorRate.String()):
		errorChart := ErrorRateChart{BaseChart: baseChart}
		getChart, err := errorChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	case strings.ToLower(pb.ChartType_ErrorCount.String()):
		errCountChart := ErrorCountChart{BaseChart: baseChart}
		getChart, err := errCountChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	case strings.ToLower(pb.ChartType_SlowCount.String()):
		slowCountChart := SlowCountChart{BaseChart: baseChart}
		getChart, err := slowCountChart.GetChart(ctx)
		if err != nil {
			return nil, err
		}
		return getChart, err
	default:
		return nil, nil
	}
}
