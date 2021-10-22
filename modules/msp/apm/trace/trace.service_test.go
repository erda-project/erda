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
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/common"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/debug"
	"github.com/erda-project/erda/modules/msp/apm/trace/core/query"
	"github.com/erda-project/erda/modules/msp/apm/trace/db"
)

func Test_traceService_GetSpans(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetSpansRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetSpansResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetSpansRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetSpansResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetSpans(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetSpans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetSpans() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_GetTraces(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTracesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTracesResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTracesRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTracesResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetTraces(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetTraces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetTraces() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_GetTraceQueryConditions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTraceQueryConditionsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTraceQueryConditionsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTraceQueryConditionsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTraceQueryConditionsResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetTraceQueryConditions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetTraceQueryConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetTraceQueryConditions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_GetTraceDebugHistories(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTraceDebugHistoriesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTraceDebugHistoriesResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTraceDebugHistoriesRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTraceDebugHistoriesResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetTraceDebugHistories(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetTraceDebugHistories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetTraceDebugHistories() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_GetTraceDebugByRequestID(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTraceDebugRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTraceDebugResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTraceDebugRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTraceDebugResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetTraceDebugByRequestID(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetTraceDebugByRequestID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetTraceDebugByRequestID() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_CreateTraceDebug(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateTraceDebugRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateTraceDebugResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.CreateTraceDebugRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.CreateTraceDebugResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.CreateTraceDebug(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.CreateTraceDebug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.CreateTraceDebug() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_StopTraceDebug(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.StopTraceDebugRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.StopTraceDebugResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.StopTraceDebugRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.StopTraceDebugResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.StopTraceDebug(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.StopTraceDebug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.StopTraceDebug() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_traceService_GetTraceDebugHistoryStatusByRequestID(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTraceDebugStatusByRequestIDRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTraceDebugStatusByRequestIDResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTraceDebugStatusByRequestIDRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTraceDebugStatusByRequestIDResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetTraceDebugHistoryStatusByRequestID(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetTraceDebugHistoryStatusByRequestID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetTraceDebugHistoryStatusByRequestID() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_composeTraceRequestHistory(t *testing.T) {
	key := uuid.NewV4().String()
	req := pb.CreateTraceDebugRequest{
		Method:    "GET",
		Url:       "http://erda.cloud",
		Body:      "",
		Query:     map[string]string{},
		Header:    map[string]string{},
		ScopeID:   key,
		ProjectID: "1",
	}

	queryString, err := json.Marshal(req.Query)
	if err != nil {
		return
	}
	headerString, err := json.Marshal(req.Header)
	if err != nil {
		return
	}
	bodyValid := json.Valid([]byte(req.Body))
	if req.Body != "" && !bodyValid {
		return
	}
	if req.CreateTime == "" || req.UpdateTime == "" {
		req.CreateTime = time.Now().Format(common.Layout)
		req.UpdateTime = time.Now().Format(common.Layout)
	}
	createTime, err := time.ParseInLocation(common.Layout, req.CreateTime, time.Local)
	if err != nil {
		return
	}
	updateTime, err := time.ParseInLocation(common.Layout, req.UpdateTime, time.Local)
	if err != nil {
		return
	}

	key2 := uuid.NewV4().String()
	req2 := pb.CreateTraceDebugRequest{
		Method:    "GET",
		Url:       "http://erda.cloud",
		Body:      "{'name':'test'}",
		Query:     map[string]string{},
		Header:    map[string]string{},
		ScopeID:   key2,
		ProjectID: "1",
	}

	queryString2, err := json.Marshal(req2.Query)
	if err != nil {
		return
	}
	headerString2, err := json.Marshal(req2.Header)
	if err != nil {
		return
	}
	bodyValid2 := json.Valid([]byte(req2.Body))
	if req2.Body != "" && !bodyValid2 {
		return
	}
	if req2.CreateTime == "" || req2.UpdateTime == "" {
		req2.CreateTime = time.Now().Format(common.Layout)
		req2.UpdateTime = time.Now().Format(common.Layout)
	}
	createTime2, err := time.ParseInLocation(common.Layout, req2.CreateTime, time.Local)
	if err != nil {
		return
	}
	updateTime2, err := time.ParseInLocation(common.Layout, req2.UpdateTime, time.Local)
	if err != nil {
		return
	}

	key3 := uuid.NewV4().String()
	req3 := pb.CreateTraceDebugRequest{
		Method:    "GET",
		Url:       "http://erda.cloud",
		Body:      "{fd'namdfasdfe'fasdx:fadsf'test'ad}",
		Query:     map[string]string{},
		Header:    map[string]string{},
		ScopeID:   key3,
		ProjectID: "1",
	}

	queryString3, err := json.Marshal(req3.Query)
	if err != nil {
		return
	}
	headerString3, err := json.Marshal(req3.Header)
	if err != nil {
		return
	}
	bodyValid3 := json.Valid([]byte(req3.Body))
	if req3.Body != "" && !bodyValid3 {
		return
	}
	if req3.CreateTime == "" || req3.UpdateTime == "" {
		req3.CreateTime = time.Now().Format(common.Layout)
		req3.UpdateTime = time.Now().Format(common.Layout)
	}
	createTime3, err := time.ParseInLocation(common.Layout, req3.CreateTime, time.Local)
	if err != nil {
		return
	}
	updateTime3, err := time.ParseInLocation(common.Layout, req3.UpdateTime, time.Local)
	if err != nil {
		return
	}
	h := &db.TraceRequestHistory{
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

	h2 := &db.TraceRequestHistory{
		TerminusKey:    req2.ScopeID,
		Url:            req2.Url,
		QueryString:    string(queryString2),
		Header:         string(headerString2),
		Body:           req2.Body,
		Method:         req2.Method,
		Status:         int(req2.Status),
		ResponseBody:   req2.ResponseBody,
		ResponseStatus: int(req2.ResponseCode),
		CreateTime:     createTime2,
		UpdateTime:     updateTime2,
	}

	h3 := &db.TraceRequestHistory{
		TerminusKey:    req3.ScopeID,
		Url:            req3.Url,
		QueryString:    string(queryString3),
		Header:         string(headerString3),
		Body:           req3.Body,
		Method:         req3.Method,
		Status:         int(req3.Status),
		ResponseBody:   req3.ResponseBody,
		ResponseStatus: int(req3.ResponseCode),
		CreateTime:     createTime3,
		UpdateTime:     updateTime3,
	}
	type args struct {
		req *pb.CreateTraceDebugRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *db.TraceRequestHistory
		wantErr bool
	}{
		{"case-1", args{&req}, h, false},
		{"case-2", args{&req2}, h2, false},
		{"case-3", args{&req3}, h3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := composeTraceRequestHistory(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("composeTraceRequestHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("composeTraceRequestHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want.RequestId = got.RequestId
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("composeTraceRequestHistory() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bodyCheck(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"case1", args{body: ""}, true},
		{"case2", args{body: "{\"test\":\"test\"}"}, true},
		{"case3", args{body: "s{ss}sss"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bodyCheck(tt.args.body); got != tt.want {
				t.Errorf("bodyCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_traceService_getDebugStatus(t *testing.T) {
	type fields struct {
		p                     *provider
		i18n                  i18n.Translator
		traceRequestHistoryDB *db.TraceRequestHistoryDB
	}
	type args struct {
		lang       i18n.LanguageCodes
		statusCode debug.Status
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case1", fields{p: nil, i18n: nil, traceRequestHistoryDB: nil}, args{lang: nil, statusCode: debug.Success}, ""},
		{"case2", fields{p: nil, i18n: nil, traceRequestHistoryDB: nil}, args{lang: nil, statusCode: debug.Init}, ""},
		{"case3", fields{p: nil, i18n: nil, traceRequestHistoryDB: nil}, args{lang: nil, statusCode: debug.Fail}, ""},
		{"case4", fields{p: nil, i18n: nil, traceRequestHistoryDB: nil}, args{lang: nil, statusCode: debug.Stop}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &traceService{
				p:                     tt.fields.p,
				i18n:                  tt.fields.i18n,
				traceRequestHistoryDB: tt.fields.traceRequestHistoryDB,
			}
			if got := s.getDebugStatus(tt.args.lang, tt.args.statusCode); got != tt.want {
				t.Errorf("getDebugStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_traceService_composeTraceQueryConditions(t *testing.T) {
	type fields struct {
		p                     *provider
		i18n                  i18n.Translator
		traceRequestHistoryDB *db.TraceRequestHistoryDB
	}
	type args struct {
		req *pb.GetTracesRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case-default", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID: "test-case-tenant-id",
			Status:   "trace_all",
			Limit:    100,
			Sort:     "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
		{"case-duration", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID:    "test-case-tenant-id",
			Status:      "trace_all",
			Limit:       100,
			DurationMin: 1000000,
			DurationMax: 3000000,
			Sort:        "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE trace_duration::field>$duration_min AND trace_duration::field<$duration_max AND errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
		{"case-traceId", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID: "test-case-tenant-id",
			Status:   "trace_all",
			Limit:    100,
			TraceID:  "test-case-trace-id",
			Sort:     "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE trace_id::tag=$trace_id AND errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
		{"case-httpPath", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID: "test-case-tenant-id",
			Status:   "trace_all",
			Limit:    100,
			TraceID:  "test-case-trace-id",
			HttpPath: "/api/health",
			Sort:     "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE trace_id::tag=$trace_id AND http_paths::field=$http_paths AND errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
		{"case-dubboMethod", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID:    "test-case-tenant-id",
			Status:      "trace_all",
			Limit:       100,
			TraceID:     "test-case-trace-id",
			DubboMethod: "io.terminus.xxx",
			Sort:        "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE trace_id::tag=$trace_id AND dubbo_methods::field=$dubbo_methods AND errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
		{"case-serviceName", fields{
			p:                     nil,
			i18n:                  nil,
			traceRequestHistoryDB: nil,
		}, args{req: &pb.GetTracesRequest{
			TenantID:    "test-case-tenant-id",
			Status:      "trace_all",
			Limit:       100,
			TraceID:     "test-case-trace-id",
			ServiceName: "apm-demo-api",
			Sort:        "100",
		}}, "SELECT start_time::field,end_time::field,service_names::field,trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE trace_id::tag=$trace_id AND service_names::field=$service_names AND errors_sum::field>=0 AND terminus_keys::field=$terminus_keys ORDER BY start_time::field DESC LIMIT 100"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &traceService{
				p:                     tt.fields.p,
				i18n:                  tt.fields.i18n,
				traceRequestHistoryDB: tt.fields.traceRequestHistoryDB,
			}
			_, want := s.composeTraceQueryConditions(tt.args.req)
			if want != tt.want {
				t.Errorf("composeTraceQueryConditions() got1 = %v, want %v", want, tt.want)
			}
		})
	}
}

func Test_getSpanProcessAnalysisDashboard(t *testing.T) {
	type args struct {
		metricType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", args{metricType: "jvm_memory"}, "span_process_analysis_java"},
		{"case2", args{metricType: "nodejs_memory"}, "span_process_analysis_nodejs"},
		{"case3", args{metricType: "xxx"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSpanProcessAnalysisDashboard(tt.args.metricType); got != tt.want {
				t.Errorf("getSpanProcessAnalysisDashboard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_traceService_getSpanCallAnalysis(t *testing.T) {
	type fields struct {
		p                     *provider
		i18n                  i18n.Translator
		traceRequestHistoryDB *db.TraceRequestHistoryDB
	}
	type args struct {
		ctx context.Context
		req *pb.GetSpanDashboardsRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case1", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "http_client"}}, "span_call_analysis_http_client"},
		{"case2", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "http_server"}}, "span_call_analysis_http_client"},
		{"case3", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "rpc_client"}}, "span_call_analysis_rpc_client"},
		{"case4", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "rpc_server"}}, "span_call_analysis_rpc_server"},
		{"case5", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "mq_producer"}}, "span_call_analysis_mq_producer"},
		{"case6", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "mq_consumer"}}, "span_call_analysis_mq_consumer"},
		{"case7", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "cache_client"}}, "span_call_analysis_cache_client"},
		{"case8", fields{}, args{req: &pb.GetSpanDashboardsRequest{Type: "invoke_local"}}, "span_call_analysis_invoke_local"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &traceService{
				p:                     tt.fields.p,
				i18n:                  tt.fields.i18n,
				traceRequestHistoryDB: tt.fields.traceRequestHistoryDB,
			}
			got, err := s.getSpanCallAnalysis(tt.args.ctx, tt.args.req)
			if (err != nil) && got.DashboardID != tt.want {
				t.Errorf("getSpanCallAnalysis() error = %v, wantErr %v", err, tt.want)
				return
			}
		})
	}
}

func Test_traceService_handleSpanResponse(t *testing.T) {
	spans := []*pb.Span{
		{
			Id:            "34035630-b97a-4f6e-b9e8-a1ca6ec32127",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Producer/Sync/apm-demo-topic-sync-parallel",
			StartTime:     1631515503925000000,
			EndTime:       1631515503927000000,
			ParentSpanId:  "5f3fd322-e4bd-4117-8449-5027cf49ef56",
		},
		{
			Id:            "fc494cf9-2811-453f-bd9e-99b1d865b7c2",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Producer/Async/apm-demo-topic-async-parallel",
			StartTime:     1631515503928000000,
			EndTime:       1631515503932000000,
			ParentSpanId:  "5f3fd322-e4bd-4117-8449-5027cf49ef56",
		},
		{
			Id:            "14eaee91-8064-41e7-9bcb-f9e69f92d1df",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Producer/Sync/apm-demo-topic-sync-order",
			StartTime:     1631515503927000000,
			EndTime:       1631515503928000000,
			ParentSpanId:  "5f3fd322-e4bd-4117-8449-5027cf49ef56",
		},
		{
			Id:            "173e9b4f-8cc9-4d26-9914-637d8a12ba83",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Consumer/apm-demo-topic",
			StartTime:     1631515503934000000,
			EndTime:       1631515503934000000,
			ParentSpanId:  "9548c236-90b0-48ac-8fb5-75100c76e129",
		},
		{
			Id:            "fe925605-78a9-409d-a506-2c6c8aecdfc2",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Consumer/apm-demo-topic-sync-parallel",
			StartTime:     1631515503930000000,
			EndTime:       1631515503930000000,
			ParentSpanId:  "34035630-b97a-4f6e-b9e8-a1ca6ec32127",
		},
		{
			Id:            "97a6d79b-c7d8-4818-83de-47fad413dcf4",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Producer/Async/apm-demo-topic-async-order",
			StartTime:     1631515503929000000,
			EndTime:       1631515503938000000,
			ParentSpanId:  "5f3fd322-e4bd-4117-8449-5027cf49ef56",
		},
		{
			Id:            "9548c236-90b0-48ac-8fb5-75100c76e129",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Producer/Sync/apm-demo-topic",
			StartTime:     1631515503923000000,
			EndTime:       1631515503925000000,
			ParentSpanId:  "5f3fd322-e4bd-4117-8449-5027cf49ef56",
		}, {

			Id:            "5f3fd322-e4bd-4117-8449-5027cf49ef56",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "GET http://ss-api.test.terminus.io/api/rocketmq/send",
			StartTime:     1631515503922000000,
			EndTime:       1631515503929000000,
			ParentSpanId:  "",
		},
		{
			Id:            "7aa7a7fc-9756-4b3a-b0b2-ecac8230242f",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Consumer/apm-demo-topic-async-order",
			StartTime:     1631515503940000000,
			EndTime:       1631515503940000000,
			ParentSpanId:  "97a6d79b-c7d8-4818-83de-47fad413dcf4",
		},
		{
			Id:            "564acae7-f22f-40ed-9049-a5c6a62d59d2",
			TraceId:       "95606e4b-92da-4f9e-8f64-4e241e30157f",
			OperationName: "RocketMQ/Consumer/apm-demo-topic-async-parallel",
			StartTime:     1631515503935000000,
			EndTime:       1631515503935000000,
			ParentSpanId:  "fc494cf9-2811-453f-bd9e-99b1d865b7c2",
		},
	}

	tree := map[string]*pb.Span{}
	for _, v := range spans {
		tree[v.Id] = v
	}
	type fields struct {
		p                     *provider
		i18n                  i18n.Translator
		traceRequestHistoryDB *db.TraceRequestHistoryDB
	}
	type args struct {
		spanTree query.SpanTree
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *pb.GetSpansResponse
	}{
		{"case1", fields{}, args{spanTree: tree}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &traceService{
				p:                     tt.fields.p,
				i18n:                  tt.fields.i18n,
				traceRequestHistoryDB: tt.fields.traceRequestHistoryDB,
			}
			got, _ := s.handleSpanResponse(tt.args.spanTree)
			if got.ServiceCount != 1 {
				t.Errorf("handleSpanResponse() got = %v, want %v", got.ServiceCount, 1)
			}

			for _, span := range got.Spans {
				if span.SelfDuration < 0 || span.Duration < 0 {
					t.Errorf("duration (%v) or selfDuration (%v) < 0", span.Duration, span.SelfDuration)
				}
			}
		})
	}
}

func Test_traceService_GetSpanEvents(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.SpanEventRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.SpanEventResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.trace.TraceService",
		//			`
		//erda.msp.apm.trace:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTracesRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTracesResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TraceServiceServer)
			got, err := srv.GetSpanEvents(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("traceService.GetSpanEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("traceService.GetSpanEvents() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
