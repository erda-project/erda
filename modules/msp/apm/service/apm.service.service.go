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

package service

import (
	context "context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	pb "github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type apmServiceService struct {
	p *provider
}

func (s *apmServiceService) GetServices(ctx context.Context, req *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	if req.TenantId == "" {
		return nil, errors.NewMissingParameterError("tenantId")
	}
	if req.PageNo <= 0 {
		req.PageNo = 1
	}
	if req.PageSize < 15 {
		req.PageSize = 15
	}
	if req.PageSize > 60 {
		req.PageSize = 60
	}

	// default get time: 1 day.
	d, _ := time.ParseDuration("-24h")
	start := time.Now().Add(d).UnixNano() / 1e6
	end := time.Now().UnixNano() / 1e6

	// get services list
	statement := fmt.Sprintf("SELECT service_id::tag,service_name::tag,service_agent_platform::tag,max(timestamp) FROM application_service_node "+
		"WHERE $condition GROUP BY service_id::tag ORDER BY max(timestamp) DESC LIMIT %v OFFSET %v", req.PageSize, (req.PageNo-1)*req.PageSize)
	condition := " terminus_key::tag=$terminus_key "
	queryParams := map[string]*structpb.Value{
		"terminus_key": structpb.NewStringValue(req.TenantId),
	}
	if req.ServiceName != "" {
		condition += " AND service_name::tag=~/.*" + req.ServiceName + ".*/ "
	}
	statement = strings.ReplaceAll(statement, "$condition", condition)
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(start, 10),
		End:       strconv.FormatInt(end, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := s.p.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	services := make([]*pb.Service, 0, 10)
	rows := response.Results[0].Series[0].Rows
	for _, row := range rows {
		service := new(pb.Service)
		service.Id = row.Values[0].GetStringValue()
		service.Name = row.Values[1].GetStringValue()
		service.Language = parseLanguage(row.Values[2].GetStringValue())
		service.LastHeartbeat = time.Unix(0, int64(row.Values[3].GetNumberValue())).Format("2006-01-02 15:04:05")
		services = append(services, service)
	}

	// calculate total count
	statement = "SELECT DISTINCT(service_id::tag) FROM application_service_node WHERE $condition"
	statement = strings.ReplaceAll(statement, "$condition", condition)
	countRequest := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(start, 10),
		End:       strconv.FormatInt(end, 10),
		Statement: statement,
		Params:    queryParams,
	}
	countResponse, err := s.p.Metric.QueryWithInfluxFormat(ctx, countRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	total := int64(countResponse.Results[0].Series[0].Rows[0].GetValues()[0].GetNumberValue())

	return &pb.GetServicesResponse{PageNo: req.PageNo, PageSize: req.PageSize, Total: total, List: services}, nil
}

func parseLanguage(platform string) (language commonpb.Language) {
	lowerPlatform := strings.ToLower(platform)
	language = commonpb.Language_unknown
	if strings.Contains(lowerPlatform, "java") || strings.Contains(lowerPlatform, "jdk") {
		language = commonpb.Language_java
	} else if strings.Contains(lowerPlatform, "nodejs") {
		language = commonpb.Language_nodejs
	} else if strings.Contains(lowerPlatform, "python") || strings.Contains(lowerPlatform, "py") {
		language = commonpb.Language_python
	} else if strings.Contains(lowerPlatform, "golang") || strings.Contains(lowerPlatform, "go") {
		language = commonpb.Language_golang
	} else if strings.Contains(lowerPlatform, "c") && !(strings.Contains(lowerPlatform, "cpp") || strings.Contains(lowerPlatform, "c++")) && (!strings.Contains(lowerPlatform, "c#")) {
		language = commonpb.Language_c
	} else if strings.Contains(lowerPlatform, "cpp") || strings.Contains(lowerPlatform, "c++") {
		language = commonpb.Language_cpp
	} else if strings.Contains(lowerPlatform, "php") {
		language = commonpb.Language_php
	} else if strings.Contains(lowerPlatform, ".net") {
		language = commonpb.Language_dotnet
	} else if strings.Contains(lowerPlatform, "c#") {
		language = commonpb.Language_csharp
	}
	return
}

func (s *apmServiceService) GetServiceAnalyzerOverview(ctx context.Context, req *pb.GetServiceAnalyzerOverviewRequest) (*pb.GetServiceAnalyzerOverviewResponse, error) {
	if req.TenantId == "" {
		return nil, errors.NewMissingParameterError("tenantId")
	}
	if req.ServiceId == "" {
		return nil, errors.NewMissingParameterError("serviceId")
	}

	// default get time: 1 h.
	h, _ := time.ParseDuration("-1h")
	start := time.Now().Add(h).UnixNano() / 1e6
	end := time.Now().UnixNano() / 1e6

	statement := "SELECT sum(count_sum::field),sum(elapsed_sum::field),sum(errors_sum::field)" +
		"FROM application_http_service,application_rpc_service,application_db_service,application_cache_service,application_mq_service " +
		"WHERE target_terminus_key::tag=$terminus_key AND target_service_id::tag=$service_id GROUP BY time(1m)"
	queryParams := map[string]*structpb.Value{
		"terminus_key": structpb.NewStringValue(req.TenantId),
		"service_id":   structpb.NewStringValue(req.ServiceId),
	}
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(start, 10),
		End:       strconv.FormatInt(end, 10),
		Statement: statement,
		Params:    queryParams,
	}
	serviceCharts := make([]*pb.ServiceChart, 0, 3)
	response, err := s.p.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	qpsCharts := make([]*pb.Chart, 0, 10)
	durationCharts := make([]*pb.Chart, 0, 10)
	errorRateCharts := make([]*pb.Chart, 0, 10)

	rows := response.Results[0].Series[0].Rows

	var (
		countSum      int64
		durationSum   int64
		errorCountSum int64
		avgDuration   float32
		errorRate     float32
	)
	for _, row := range rows {
		qpsChart := new(pb.Chart)
		durationChart := new(pb.Chart)
		errorRateChart := new(pb.Chart)
		date := row.Values[0].GetStringValue()
		parse, err := time.ParseInLocation("2006-01-02T15:04:05Z", date, time.Local)
		if err != nil {
			return nil, err
		}
		timestamp := parse.UnixNano() / int64(time.Millisecond)
		qpsChart.Timestamp = timestamp
		durationChart.Timestamp = timestamp
		errorRateChart.Timestamp = timestamp

		count := int64(row.Values[1].GetNumberValue())
		duration := int64(row.Values[2].GetNumberValue())
		errorCount := int64(row.Values[3].GetNumberValue())
		qpsChart.Value = float32(count) / 60
		if count != 0 {
			durationChart.Value = float32(duration / count)
			errorRateChart.Value = float32(errorCount / count)
		}

		countSum += count
		durationSum += duration
		errorCount += errorCount

		qpsCharts = append(qpsCharts, qpsChart)
		durationCharts = append(durationCharts, durationChart)
		errorRateCharts = append(errorRateCharts, errorRateChart)
	}

	if countSum != 0 {
		avgDuration = float32(durationSum / countSum)
		errorRate = float32(errorCountSum / countSum)
	}

	// QPS Chart
	serviceCharts = append(serviceCharts, &pb.ServiceChart{
		Type: pb.ChartType_QPS.String(),
		Data: float32(countSum / (60 * 60)),
		View: qpsCharts,
	})

	// Avg Duration Chart
	serviceCharts = append(serviceCharts, &pb.ServiceChart{
		Type: pb.ChartType_AvgDuration.String(),
		Data: avgDuration,
		View: durationCharts,
	})

	// Error Rate Chart
	serviceCharts = append(serviceCharts, &pb.ServiceChart{
		Type: pb.ChartType_ErrorRate.String(),
		Data: errorRate,
		View: errorRateCharts,
	})

	return &pb.GetServiceAnalyzerOverviewResponse{List: serviceCharts}, nil
}
