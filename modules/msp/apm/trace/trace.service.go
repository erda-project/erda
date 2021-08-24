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

package trace

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
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
	mathpkg "github.com/erda-project/erda/pkg/math"
)

type traceService struct {
	p                     *provider
	i18n                  i18n.Translator
	traceRequestHistoryDB *db.TraceRequestHistoryDB
}

const layout = "2006-01-02 15:04:05"

type DebugStatus int32

const (
	DebugInit    DebugStatus = 0
	DebugSuccess DebugStatus = 1
	DebugFail    DebugStatus = 2
	DebugStop    DebugStatus = 3
)

func (s *traceService) getDebugStatus(lang i18n.LanguageCodes, statusCode DebugStatus) string {
	if lang == nil {
		return ""
	}
	switch statusCode {
	case DebugInit:
		return s.i18n.Text(lang, "waiting_for_tracing_data")
	case DebugSuccess:
		return s.i18n.Text(lang, "success_get_tracing_data")
	case DebugFail:
		return s.i18n.Text(lang, "fail_get_tracing_data")
	case DebugStop:
		return s.i18n.Text(lang, "stop_get_tracing_data")
	default:
		return ""
	}
}

type SpanTree map[string]*pb.Span

func (s *traceService) GetSpans(ctx context.Context, req *pb.GetSpansRequest) (*pb.GetSpansResponse, error) {
	if req.TraceID == "" || req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("traceId or scopeId")
	}
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 1000
	}
	iter := s.p.cassandraSession.Query("SELECT * FROM spans WHERE trace_id = ? limit ?", req.TraceID, req.Limit).Iter()
	spanTree := make(SpanTree)
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

type void struct{}

func (s *traceService) handleSpanResponse(spanTree SpanTree) (*pb.GetSpansResponse, error) {
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
	services := map[string]void{}
	for _, span := range spanTree {
		services[span.Tags["service_name"]] = void{}
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
		spans = append(spans, span)
	}

	serviceCount := int64(len(services))

	return &pb.GetSpansResponse{Spans: spans, ServiceCount: serviceCount, Depth: depth, Duration: mathpkg.AbsInt64(traceEndTime - traceStartTime), SpanCount: spanCount}, nil
}

func calculateDepth(depth int64, span *pb.Span, spanTree SpanTree) int64 {
	if span.ParentSpanId != "" && spanTree[span.ParentSpanId] != nil {
		depth += 1
		calculateDepth(depth, spanTree[span.ParentSpanId], spanTree)
	}
	return depth
}

func (s *traceService) GetSpanCount(ctx context.Context, traceID string) (int64, error) {
	count := 0
	s.p.cassandraSession.Query("SELECT COUNT(trace_id) FROM spans WHERE trace_id = ?", traceID).Iter().Scan(&count)
	return int64(count), nil
}

func (s *traceService) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	if req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("scopeId")
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if req.EndTime <= 0 || req.StartTime <= 0 {
		req.EndTime = time.Now().UnixNano() / 1e6
		h, _ := time.ParseDuration("-1h")
		req.StartTime = time.Now().Add(h).UnixNano() / 1e6
	}
	metricsParams := url.Values{}
	metricsParams.Set("start", strconv.FormatInt(req.StartTime, 10))
	metricsParams.Set("end", strconv.FormatInt(req.EndTime, 10))

	queryParams := make(map[string]*structpb.Value)
	queryParams["terminus_keys"] = structpb.NewStringValue(req.ScopeID)
	var where bytes.Buffer
	if req.ApplicationID > 0 {
		queryParams["applications_ids"] = structpb.NewStringValue(strconv.FormatInt(req.ApplicationID, 10))
		where.WriteString("applications_ids::field=$applications_ids AND ")
	}
	if req.TraceID != "" {
		queryParams["trace_id"] = structpb.NewStringValue(req.TraceID)
		where.WriteString("trace_id::tag=$trace_id AND ")
	}

	// -1 error, 0 both, 1 success
	if req.Status == 1 {
		where.WriteString("errors_sum::field=0 AND")
	} else if req.Status == 0 {
		where.WriteString("errors_sum::field>=0 AND")
	} else if req.Status == -1 {
		where.WriteString("errors_sum::field>0 AND")
	} else {
		return nil, errors.NewParameterTypeError("status just -1,0,1")
	}

	statement := fmt.Sprintf("SELECT start_time::field,end_time::field,components::field,"+
		"trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE %s terminus_keys::field=$terminus_keys "+
		"ORDER BY start_time::field DESC LIMIT %s", where.String(), strconv.FormatInt(req.Limit, 10))

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
	traces := make([]*pb.Trace, 0, len(rows))
	for _, row := range rows {
		var trace pb.Trace
		values := row.Values
		trace.StartTime = int64(values[0].GetNumberValue() / 1e6)
		trace.Elapsed = math.Abs(values[1].GetNumberValue() - values[0].GetNumberValue())
		for _, serviceName := range values[2].GetListValue().Values {
			trace.Services = append(trace.Services, serviceName.GetStringValue())
		}
		trace.Id = values[3].GetStringValue()
		traces = append(traces, &trace)
	}

	return &pb.GetTracesResponse{Data: traces}, nil
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
		StatusName: s.getDebugStatus(apis.Language(ctx), DebugStatus(insertHistory.Status)),
		ScopeID:    insertHistory.TerminusKey,
	}

	response, err := s.sendHTTPRequest(err, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	responseCode := response.StatusCode

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.p.Log.Error("http response close fail.")
		}
	}(response.Body)

	_, err = s.traceRequestHistoryDB.UpdateDebugResponseByRequestID(req.ScopeID, req.RequestID, responseCode)
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
	if req.CreateTime == "" || req.UpdateTime == "" {
		req.CreateTime = time.Now().Format(layout)
		req.UpdateTime = time.Now().Format(layout)
	}
	createTime, err := time.ParseInLocation(layout, req.CreateTime, time.Local)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateTime, err := time.ParseInLocation(layout, req.UpdateTime, time.Local)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	history := &db.TraceRequestHistory{
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
	_, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(req.ScopeID, req.RequestID, int(DebugFail))
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
		StatusName: s.getDebugStatus(apis.Language(ctx), DebugStatus(dbHistory.Status)),
		ScopeID:    dbHistory.TerminusKey,
	}

	if DebugStatus(dbHistory.Status) == DebugInit {
		exist, err := s.isExistSpan(ctx, req.RequestID)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		if exist {
			info, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(dbHistory.TerminusKey, dbHistory.RequestId, int(DebugSuccess))
			if err != nil {
				return nil, errors.NewInternalServerError(err)
			}
			statusInfo.Status = int32(info.Status)
			statusInfo.StatusName = s.getDebugStatus(apis.Language(ctx), DebugStatus(info.Status))
		} else {
			// If trace data is not obtained within 20 minutes, it is considered that the writing of trace data has failed.
			if (time.Now().UnixNano()-dbHistory.UpdateTime.UnixNano())/1e9 > 20*60 {
				info, err := s.traceRequestHistoryDB.UpdateDebugStatusByRequestID(dbHistory.TerminusKey, dbHistory.RequestId, int(DebugFail))
				if err != nil {
					return nil, errors.NewInternalServerError(err)
				}
				statusInfo.Status = int32(info.Status)
				statusInfo.StatusName = s.getDebugStatus(apis.Language(ctx), DebugStatus(info.Status))
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
		StatusName:   s.getDebugStatus(language, DebugStatus(dbHistory.Status)),
		ResponseCode: int32(dbHistory.ResponseStatus),
		ResponseBody: dbHistory.ResponseBody,
		Method:       dbHistory.Method,
		CreateTime:   dbHistory.CreateTime.Format(layout),
		UpdateTime:   dbHistory.UpdateTime.Format(layout),
	}, nil
}
