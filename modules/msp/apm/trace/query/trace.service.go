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

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/pkg/set"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/msp/apm/trace"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/common"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/debug"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/query"
	"github.com/erda-project/erda/modules/msp/apm/trace/db"
	"github.com/erda-project/erda/modules/msp/apm/trace/storage"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
	mathpkg "github.com/erda-project/erda/pkg/math"
)

type traceService struct {
	p                     *provider
	i18n                  i18n.Translator
	traceRequestHistoryDB *db.TraceRequestHistoryDB
	StorageReader         storage.Storage
}

var EventFieldSet = set.NewSet("error", "stack", "event", "message", "error_kind", "error_object")

func (s *traceService) getDebugStatus(lang i18n.LanguageCodes, statusCode debug.Status) string {
	if lang == nil {
		return ""
	}
	switch statusCode {
	case debug.Init:
		return s.i18n.Text(lang, "waiting_for_tracing_data")
	case debug.Success:
		return s.i18n.Text(lang, "success_get_tracing_data")
	case debug.Fail:
		return s.i18n.Text(lang, "fail_get_tracing_data")
	case debug.Stop:
		return s.i18n.Text(lang, "stop_get_tracing_data")
	default:
		return ""
	}
}

func (s *traceService) GetSpans(ctx context.Context, req *pb.GetSpansRequest) (*pb.GetSpansResponse, error) {
	if req.TraceID == "" || req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("traceId or scopeId")
	}
	if req.Limit <= 0 || req.Limit > 10000 {
		req.Limit = 10000
	}
	spanTree := make(query.SpanTree)

	// do es query
	it, err := s.StorageReader.Iterator(ctx, &storage.Selector{
		TraceId: req.TraceID,
	})
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer it.Close()

	items, err := fetchSpanFromES(it, true, int(req.GetLimit()))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	for _, value := range items {
		var span pb.Span
		span.Id = value.SpanId
		span.TraceId = value.TraceId
		span.OperationName = value.OperationName
		span.ParentSpanId = value.ParentSpanId
		span.StartTime = value.StartTime
		span.EndTime = value.EndTime
		span.Tags = value.Tags
		spanTree[span.Id] = &span
	}

	// do cassandra query, will be removed in future
	iter := s.p.cassandraSession.Session().Query("SELECT * FROM spans WHERE trace_id = ? limit ?", req.TraceID, req.Limit).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		var span pb.Span
		span.Id = row["span_id"].(string)
		span.TraceId = row["trace_id"].(string)
		span.OperationName = row["operation_name"].(string)
		span.ParentSpanId = row["parent_span_id"].(string)
		span.StartTime = row["start_time"].(int64)
		span.EndTime = row["end_time"].(int64)
		span.Tags = row["tags"].(map[string]string)
		spanTree[span.Id] = &span
	}

	response, err := s.handleSpanResponse(spanTree)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return response, nil
}

func fetchSpanFromES(it storekit.Iterator, forward bool, limit int) (list []*trace.Span, err error) {
	if forward {
		for it.Next() {
			if len(list) >= limit {
				return list, nil
			}
			span, ok := it.Value().(*trace.Span)
			if !ok {
				continue
			}
			list = append(list, span)
		}
	} else {
		for it.Prev() {
			span, ok := it.Value().(*trace.Span)
			if !ok {
				continue
			}
			if len(list) >= limit {
				return list, nil
			}
			list = append(list, span)
		}
		sort.Sort(Spans(list))
	}
	return list, it.Error()
}

type Spans []*trace.Span

func (s Spans) Len() int      { return len(s) }
func (s Spans) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Spans) Less(i, j int) bool {
	return s[i].StartTime < s[j].StartTime
}

func getSpanProcessAnalysisDashboard(metricType string) string {
	switch metricType {
	case query.JavaMemoryMetricName:
		return "span_process_analysis_java"
	case query.NodeJsMemoryMetricName:
		return "span_process_analysis_nodejs"
	default:
		return ""
	}
}

func (s *traceService) getServiceInstanceType(startTime, endTime int64, tenantId, serviceInstanceId string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, metricType := range query.ProcessMetrics {
		statement := fmt.Sprintf("SELECT terminus_key::tag FROM %s WHERE terminus_key=$terminus_key "+
			"AND service_instance_id=$service_instance_id LIMIT 1", metricType)
		queryParams := map[string]*structpb.Value{
			"terminus_key":        structpb.NewStringValue(tenantId),
			"service_instance_id": structpb.NewStringValue(serviceInstanceId),
		}

		request := &metricpb.QueryWithInfluxFormatRequest{
			Start:     strconv.FormatInt(startTime, 10),
			End:       strconv.FormatInt(endTime, 10),
			Statement: statement,
			Params:    queryParams,
		}

		response, err := s.p.Metric.QueryWithInfluxFormat(ctx, request)
		if err != nil {
			return "", errors.NewInternalServerError(err)
		}

		rows := response.Results[0].Series[0].Rows
		if len(rows) == 1 {
			return metricType, nil
		}
	}
	return "", nil
}

func (s *traceService) getSpanServiceAnalysis(ctx context.Context, req *pb.GetSpanDashboardsRequest) (*pb.SpanAnalysis, error) {
	instanceType, err := s.getServiceInstanceType(req.StartTime, req.EndTime, req.TenantID, req.ServiceInstanceID)
	if err != nil {
		return nil, err
	}
	dashboardId := getSpanProcessAnalysisDashboard(instanceType)

	return &pb.SpanAnalysis{
		DashboardID: dashboardId,
		Conditions:  []string{"service_instance_id"},
	}, nil
}

func (s *traceService) getSpanCallAnalysis(ctx context.Context, req *pb.GetSpanDashboardsRequest) (*pb.SpanAnalysis, error) {
	switch req.Type {
	case strings.ToLower(pb.SpanType_HTTP_CLIENT.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisHttpClient, Conditions: []string{"http_path", "http_method", "source_service_id"}}, nil
	case strings.ToLower(pb.SpanType_HTTP_SERVER.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisHttpServer, Conditions: []string{"http_path", "http_method", "target_service_id"}}, nil
	case strings.ToLower(pb.SpanType_RPC_CLIENT.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisRpcClient, Conditions: []string{"dubbo_service", "dubbo_method", "source_service_id"}}, nil
	case strings.ToLower(pb.SpanType_RPC_SERVER.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisRpcServer, Conditions: []string{"dubbo_service", "dubbo_method", "target_service_id"}}, nil
	case strings.ToLower(pb.SpanType_CACHE_CLIENT.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisCacheClient, Conditions: []string{"db_statement", "source_service_id"}}, nil
	case strings.ToLower(pb.SpanType_MQ_PRODUCER.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisMqProducer, Conditions: []string{"span_kind", "target_service_id", "message_bus_destination"}}, nil
	case strings.ToLower(pb.SpanType_MQ_CONSUMER.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisMqConsumer, Conditions: []string{"span_kind", "source_service_id", "message_bus_destination"}}, nil
	case strings.ToLower(pb.SpanType_INVOKE_LOCAL.String()):
		return &pb.SpanAnalysis{DashboardID: common.CallAnalysisInvokeLocal, Conditions: []string{"service_id"}}, nil
	default:
		return nil, errors.NewNotFoundError(fmt.Sprintf("span type (%s)", req.Type))
	}
}

func (s *traceService) GetSpanDashboards(ctx context.Context, req *pb.GetSpanDashboardsRequest) (*pb.GetSpanDashboardsResponse, error) {
	// call details analysis
	callAnalysis, err := s.getSpanCallAnalysis(ctx, req)
	if err != nil {
		s.p.Log.Error(err)
	}
	// relate service analysis
	serviceAnalysis, err := s.getSpanServiceAnalysis(ctx, req)
	if err != nil {
		s.p.Log.Error(err)
	}
	return &pb.GetSpanDashboardsResponse{
		CallAnalysis:    callAnalysis,
		ServiceAnalysis: serviceAnalysis,
	}, nil
}

func (s *traceService) handleSpanResponse(spanTree query.SpanTree) (*pb.GetSpansResponse, error) {
	var (
		spans          []*pb.Span
		traceStartTime int64 = 0
		traceEndTime   int64 = 0
		depth          int64 = 0
	)

	spanCount := int64(len(spanTree))
	if spanCount > 0 {
		depth = 1
	}
	services := map[string]common.Void{}
	for id, span := range spanTree {
		services[span.Tags["service_name"]] = common.Void{}
		if span.ParentSpanId == span.Id {
			span.ParentSpanId = ""
		}
		tempDepth := calculateDepth(depth, span, spanTree)
		if tempDepth > depth {
			depth = tempDepth
		}
		if traceStartTime == 0 || traceStartTime > span.StartTime {
			traceStartTime = span.StartTime
		}
		if traceEndTime == 0 || traceEndTime < span.EndTime {
			traceEndTime = span.EndTime
		}
		span.Duration = mathpkg.AbsInt64(span.EndTime - span.StartTime)
		span.SelfDuration = mathpkg.AbsInt64(span.Duration - childSpanDuration(id, spanTree))
		spans = append(spans, span)
	}

	serviceCount := int64(len(services))

	return &pb.GetSpansResponse{Spans: spans, ServiceCount: serviceCount, Depth: depth, Duration: mathpkg.AbsInt64(traceEndTime - traceStartTime), SpanCount: spanCount}, nil
}

func childSpanDuration(id string, spanTree query.SpanTree) int64 {
	duration := int64(0)
	for _, span := range spanTree {
		if span.ParentSpanId == id {
			duration += span.EndTime - span.StartTime
		}
	}
	return duration
}

func calculateDepth(depth int64, span *pb.Span, spanTree query.SpanTree) int64 {
	if span.ParentSpanId == span.Id {
		return 0
	}
	if span.ParentSpanId != "" && spanTree[span.ParentSpanId] != nil {
		depth += 1
		calculateDepth(depth, spanTree[span.ParentSpanId], spanTree)
	}
	return depth
}

func (s *traceService) GetSpanCount(ctx context.Context, traceID string) (int64, error) {
	cassandraCount := 0
	s.p.cassandraSession.Session().Query("SELECT COUNT(trace_id) FROM spans WHERE trace_id = ?", traceID).Iter().Scan(&cassandraCount)

	elasticsearchCount := s.StorageReader.Count(ctx, traceID)

	return int64(cassandraCount) + elasticsearchCount, nil
}

func (s *traceService) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	if req.TenantID == "" {
		return nil, errors.NewMissingParameterError("tenantId")
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if (req.DurationMin != 0 && req.DurationMax == 0) || (req.DurationMax != 0 && req.DurationMin == 0) {
		return nil, errors.NewInvalidParameterError("duration", "missing min or max duration")
	}
	if req.DurationMax != 0 && req.DurationMin != 0 && req.DurationMax <= req.DurationMin {
		return nil, errors.NewInvalidParameterError("duration", "duration min <= duration max")
	}
	if req.EndTime <= 0 || req.StartTime <= 0 {
		req.EndTime = time.Now().UnixNano() / 1e6
		h, _ := time.ParseDuration("-1h")
		req.StartTime = time.Now().Add(h).UnixNano() / 1e6
	}

	queryParams, statement := s.composeTraceQueryConditions(req)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(req.StartTime, 10),
		End:       strconv.FormatInt(req.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}

	response, err := s.p.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	rows := response.Results[0].Series[0].Rows
	traces := s.handleTracesResponse(rows)

	return &pb.GetTracesResponse{Data: traces}, nil
}

func (s *traceService) handleTracesResponse(rows []*metricpb.Row) []*pb.Trace {
	traces := make([]*pb.Trace, 0, len(rows))
	for _, row := range rows {
		var trace pb.Trace
		values := row.Values
		trace.StartTime = int64(values[0].GetNumberValue() / 1e6)
		trace.Duration = math.Abs(values[1].GetNumberValue() - values[0].GetNumberValue())
		for _, serviceName := range values[2].GetListValue().Values {
			trace.Services = append(trace.Services, serviceName.GetStringValue())
		}
		trace.Id = values[3].GetStringValue()
		traces = append(traces, &trace)
	}
	return traces
}

func (s *traceService) composeTraceQueryConditions(req *pb.GetTracesRequest) (map[string]*structpb.Value, string) {
	metricsParams := url.Values{}
	metricsParams.Set("start", strconv.FormatInt(req.StartTime, 10))
	metricsParams.Set("end", strconv.FormatInt(req.EndTime, 10))

	queryParams := make(map[string]*structpb.Value)
	queryParams["terminus_keys"] = structpb.NewStringValue(req.TenantID)

	var where bytes.Buffer
	// trace id condition
	if req.TraceID != "" {
		queryParams["trace_id"] = structpb.NewStringValue(req.TraceID)
		where.WriteString("trace_id::tag=$trace_id AND ")
	}

	if req.ServiceName != "" {
		queryParams["service_names"] = structpb.NewStringValue(req.ServiceName)
		where.WriteString("service_names::field=$service_names AND ")
	}

	if req.DubboMethod != "" {
		queryParams["dubbo_methods"] = structpb.NewStringValue(req.DubboMethod)
		where.WriteString("dubbo_methods::field=$dubbo_methods AND ")
	}

	if req.HttpPath != "" {
		queryParams["http_paths"] = structpb.NewStringValue(req.HttpPath)
		where.WriteString("http_paths::field=$http_paths AND ")
	}

	if req.DurationMin > 0 && req.DurationMax > 0 && req.DurationMin < req.DurationMax {
		queryParams["duration_min"] = structpb.NewNumberValue(float64(req.DurationMin))
		queryParams["duration_max"] = structpb.NewNumberValue(float64(req.DurationMax))
		where.WriteString("trace_duration::field>$duration_min AND trace_duration::field<$duration_max AND ")
	}

	// trace status condition
	where.WriteString(s.traceStatusConditionStrategy(req.Status))
	// sort condition
	sort := s.sortConditionStrategy(req.Sort)

	statement := fmt.Sprintf("SELECT start_time::field,end_time::field,service_names::field,"+
		"trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE %s terminus_keys::field=$terminus_keys "+
		"%s LIMIT %s", where.String(), sort, strconv.FormatInt(req.Limit, 10))
	return queryParams, statement
}

func (s *traceService) traceStatusConditionStrategy(traceStatus string) string {
	switch traceStatus {
	case strings.ToLower(pb.TraceStatusCondition_TRACE_SUCCESS.String()):
		return "errors_sum::field=0 AND"
	case strings.ToLower(pb.TraceStatusCondition_TRACE_ALL.String()):
		return "errors_sum::field>=0 AND"
	case strings.ToLower(pb.TraceStatusCondition_TRACE_ERROR.String()):
		return "errors_sum::field>0 AND"
	default:
		return "errors_sum::field>=0 AND"
	}
}

func (s *traceService) sortConditionStrategy(sort string) string {
	switch sort {
	case strings.ToLower(pb.SortCondition_TRACE_TIME_DESC.String()):
		return "ORDER BY start_time::field DESC"
	case strings.ToLower(pb.SortCondition_TRACE_TIME_ASC.String()):
		return "ORDER BY start_time::field ASC"
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_DESC.String()):
		return "ORDER BY trace_duration::field DESC"
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_ASC.String()):
		return "ORDER BY trace_duration::field ASC"
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_DESC.String()):
		return "ORDER BY span_count::field DESC"
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_ASC.String()):
		return "ORDER BY span_count::field ASC"
	default:
		return "ORDER BY start_time::field DESC"
	}
}

func (s *traceService) GetTraceDebugHistories(ctx context.Context, req *pb.GetTraceDebugHistoriesRequest) (*pb.GetTraceDebugHistoriesResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	if req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("scopeId")
	}

	histories, err := s.traceRequestHistoryDB.QueryHistoriesByScopeID(req.ScopeID, time.Now(), req.Limit)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	td := pb.TraceDebug{}
	var traceDebugHistories []*pb.TraceDebugHistory
	for _, history := range histories {
		debugHistory, err := s.convertToTraceDebugHistory(ctx, history)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		traceDebugHistories = append(traceDebugHistories, debugHistory)
	}
	td.History = traceDebugHistories
	td.Limit = int32(req.Limit)
	count, err := s.traceRequestHistoryDB.QueryCountByScopeID(req.ScopeID)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	td.Total = count
	return &pb.GetTraceDebugHistoriesResponse{Data: &td}, nil
}

func (s *traceService) GetTraceQueryConditions(ctx context.Context, req *pb.GetTraceQueryConditionsRequest) (*pb.GetTraceQueryConditionsResponse, error) {
	return &pb.GetTraceQueryConditionsResponse{Data: s.translateConditions(apis.Language(ctx))}, nil
}

func (s *traceService) translateConditions(lang i18n.LanguageCodes) *pb.TraceQueryConditions {
	conditions := query.DepthCopyQueryConditions()

	for _, condition := range conditions.Sort {
		condition.DisplayName = query.TranslateCondition(s.i18n, lang, condition.Key)
	}

	for _, condition := range conditions.TraceStatus {
		condition.DisplayName = query.TranslateCondition(s.i18n, lang, condition.Key)
	}

	for _, condition := range conditions.Others {
		condition.DisplayName = query.TranslateCondition(s.i18n, lang, condition.Key)
	}
	return conditions
}

func (s *traceService) GetTraceDebugByRequestID(ctx context.Context, req *pb.GetTraceDebugRequest) (*pb.GetTraceDebugResponse, error) {
	dbHistory, err := s.traceRequestHistoryDB.QueryHistoryByRequestID(req.ScopeID, req.RequestID)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	debugHistory, err := s.convertToTraceDebugHistory(ctx, dbHistory)
	return &pb.GetTraceDebugResponse{Data: debugHistory}, nil
}

func (s *traceService) CreateTraceDebug(ctx context.Context, req *pb.CreateTraceDebugRequest) (*pb.CreateTraceDebugResponse, error) {
	bodyValid := bodyCheck(req.Body)
	if !bodyValid {
		return nil, errors.NewParameterTypeError("body")
	}
	if req.Name == "" {
		req.Name = "no name"
	}

	history, err := composeTraceRequestHistory(req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	insertHistory, err := s.traceRequestHistoryDB.InsertHistory(*history)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	statusInfo := pb.TraceDebugStatus{
		RequestID:  insertHistory.RequestId,
		Status:     int32(insertHistory.Status),
		StatusName: s.getDebugStatus(apis.Language(ctx), debug.Status(insertHistory.Status)),
		ScopeID:    insertHistory.TerminusKey,
	}

	response, err := s.sendHTTPRequest(err, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	responseCode := response.StatusCode
	responseBody, err := ioutil.ReadAll(response.Body)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.p.Log.Error("http response close fail.")
		}
	}(response.Body)

	_, err = s.traceRequestHistoryDB.UpdateDebugResponseByRequestID(req.ScopeID, req.RequestID, responseCode, string(responseBody))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.CreateTraceDebugResponse{Data: &statusInfo}, nil
}

func bodyCheck(body string) bool {
	if body == "" {
		return true
	}
	return json.Valid([]byte(body))
}

func composeTraceRequestHistory(req *pb.CreateTraceDebugRequest) (*db.TraceRequestHistory, error) {
	req.RequestID = uuid.NewV4().String()

	queryString, err := json.Marshal(req.Query)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	headerString, err := json.Marshal(req.Header)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if req.CreateTime == "" {
		req.CreateTime = time.Now().Format(common.Layout)
	}
	createTime, err := time.ParseInLocation(common.Layout, req.CreateTime, time.Local)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req.UpdateTime = time.Now().Format(common.Layout)
	updateTime, err := time.ParseInLocation(common.Layout, req.UpdateTime, time.Local)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	history := &db.TraceRequestHistory{
		Name:           req.Name,
		RequestId:      req.RequestID,
		TerminusKey:    req.ScopeID,
		Url:            req.Url,
		QueryString:    string(queryString),
		Header:         string(headerString),
		Body:           req.Body,
		Method:         req.Method,
		Status:         int(req.Status),
		ResponseBody:   req.ResponseBody,
		ResponseStatus: int(req.ResponseCode),
		CreateTime:     createTime,
		UpdateTime:     updateTime,
	}
	return history, nil
}

func (s *traceService) sendHTTPRequest(err error, req *pb.CreateTraceDebugRequest) (*http.Response, error) {
	client := &http.Client{}
	params := ioutil.NopCloser(strings.NewReader(req.Body))
	request, err := http.NewRequest(req.Method, req.Url, params)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	for k, v := range req.Header {
		request.Header.Set(k, v)
	}
	s.tracing(request, req)
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return response, nil
}

func (s *traceService) tracing(request *http.Request, req *pb.CreateTraceDebugRequest) {
	request.Header.Set("terminus-request-id", req.RequestID)
	request.Header.Set("terminus-request-sampled", "true")
}

func (s *traceService) StopTraceDebug(ctx context.Context, req *pb.StopTraceDebugRequest) (*pb.StopTraceDebugResponse, error) {
	_, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(req.ScopeID, req.RequestID, int(debug.Fail))
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return nil, nil
}

func (s *traceService) isExistSpan(ctx context.Context, requestID string) (bool, error) {
	count, err := s.GetSpanCount(ctx, requestID)
	if err != nil {
		return false, errors.NewInternalServerError(err)
	}
	if count < 0 {
		return false, errors.NewInternalServerError(err)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (s *traceService) GetTraceDebugHistoryStatusByRequestID(ctx context.Context, req *pb.GetTraceDebugStatusByRequestIDRequest) (*pb.GetTraceDebugStatusByRequestIDResponse, error) {
	dbHistory, err := s.traceRequestHistoryDB.QueryHistoryByRequestID(req.ScopeID, req.RequestID)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if dbHistory == nil {
		return nil, errors.NewNotFoundError("trace debug history not found.")
	}

	statusInfo := pb.TraceDebugStatus{
		RequestID:  dbHistory.RequestId,
		Status:     int32(dbHistory.Status),
		StatusName: s.getDebugStatus(apis.Language(ctx), debug.Status(dbHistory.Status)),
		ScopeID:    dbHistory.TerminusKey,
	}

	if debug.Status(dbHistory.Status) == debug.Init {
		exist, err := s.isExistSpan(ctx, req.RequestID)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		if exist {
			info, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(dbHistory.TerminusKey, dbHistory.RequestId, int(debug.Success))
			if err != nil {
				return nil, errors.NewInternalServerError(err)
			}
			statusInfo.Status = int32(info.Status)
			statusInfo.StatusName = s.getDebugStatus(apis.Language(ctx), debug.Status(info.Status))
		} else {
			// If trace data is not obtained within 20 minutes, it is considered that the writing of trace data has failed.
			if (time.Now().UnixNano()-dbHistory.UpdateTime.UnixNano())/1e9 > 20*60 {
				info, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(dbHistory.TerminusKey, dbHistory.RequestId, int(debug.Fail))
				if err != nil {
					return nil, errors.NewInternalServerError(err)
				}
				statusInfo.Status = int32(info.Status)
				statusInfo.StatusName = s.getDebugStatus(apis.Language(ctx), debug.Status(info.Status))
			}
		}
	}

	return &pb.GetTraceDebugStatusByRequestIDResponse{Data: &statusInfo}, nil
}

func (s *traceService) convertToTraceDebugHistory(ctx context.Context, dbHistory *db.TraceRequestHistory) (*pb.TraceDebugHistory, error) {
	language := apis.Language(ctx)

	query := make(map[string]string)
	if dbHistory.QueryString != "" {
		err := json.Unmarshal([]byte(dbHistory.QueryString), &query)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}
	headers := make(map[string]string)
	if dbHistory.Header != "" {
		err := json.Unmarshal([]byte(dbHistory.Header), &headers)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}
	return &pb.TraceDebugHistory{
		RequestID:    dbHistory.RequestId,
		ScopeID:      dbHistory.TerminusKey,
		Url:          dbHistory.Url,
		Query:        query,
		Header:       headers,
		Body:         dbHistory.Body,
		Status:       int32(dbHistory.Status),
		StatusName:   s.getDebugStatus(language, debug.Status(dbHistory.Status)),
		ResponseCode: int32(dbHistory.ResponseStatus),
		ResponseBody: dbHistory.ResponseBody,
		Method:       dbHistory.Method,
		CreateTime:   dbHistory.CreateTime.Format(common.Layout),
		UpdateTime:   dbHistory.UpdateTime.Format(common.Layout),
	}, nil
}

func (s *traceService) GetSpanEvents(ctx context.Context, req *pb.SpanEventRequest) (*pb.SpanEventResponse, error) {
	startTime, endTime := s.getSpanEventQueryTime(req)
	req.StartTime = req.StartTime - int64((time.Minute*15)/time.Millisecond)
	statement := "select * from apm_span_event where span_id = $span_id order by timestamp asc limit 1000"
	queryParams := map[string]*structpb.Value{
		"span_id": structpb.NewStringValue(req.SpanID),
	}
	queryCtx, _ := context.WithTimeout(ctx, time.Minute)
	queryRequest := &metricpb.QueryWithTableFormatRequest{
		Start:     strconv.FormatInt(startTime, 10),
		End:       strconv.FormatInt(endTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := s.p.Metric.QueryWithTableFormat(queryCtx, queryRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	table := response.Data
	events := s.handleSpanEventResponse(table)
	return &pb.SpanEventResponse{SpanEvents: events}, nil
}

func (s *traceService) getSpanEventQueryTime(req *pb.SpanEventRequest) (int64, int64) {
	if req.StartTime <= 0 {
		req.StartTime = time.Now().Add(-time.Minute*15).UnixNano() / 1e6
	}
	return req.StartTime - int64((time.Minute*15)/time.Millisecond), req.StartTime + int64((time.Minute*15)/time.Millisecond)
}

func (s *traceService) handleSpanEventResponse(table *metricpb.TableResult) []*pb.SpanEvent {
	spanEvents := make([]*pb.SpanEvent, 0)
	eventNames := make(map[string]string)
	for _, col := range table.Cols {
		if col.Flag == "tag" {
			key := strings.Replace(col.Key, "::tag", "", -1)
			if EventFieldSet.Contains(key) {
				eventNames[key] = col.Key
			}
		}
	}
	for _, data := range table.Data {
		timestamp := int64(data.Values["timestamp"].GetNumberValue())
		events := make(map[string]string)
		for key, value := range eventNames {
			events[key] = data.Values[value].GetStringValue()
		}
		event := &pb.SpanEvent{
			Timestamp: timestamp,
			Events:    events,
		}
		spanEvents = append(spanEvents, event)
	}
	return spanEvents
}
