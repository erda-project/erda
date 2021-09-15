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

package jaeger

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-oap-thirdparty-protocol/jaeger-thrift/jaeger"
	jaegerpb "github.com/erda-project/erda-proto-go/oap/collector/receiver/jaeger/pb"
	tracing "github.com/erda-project/erda-proto-go/oap/trace/pb"
	"github.com/erda-project/erda/modules/oap/collector/receivers/common"
)

var (
	acceptedThriftFormats = map[string]struct{}{
		"application/x-thrift":                 {},
		"application/vnd.apache.thrift.binary": {},
	}

	JAEGER_MSP_ENV_ID    = "msp.env.id"
	JAEGER_MSP_AK_ID     = "msp.ak.id"
	JAEGER_MSP_AK_SECRET = "msp.ak.secret"
)

func injectCtx(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, entity interface{}) (interface{}, error) {
		header := transport.ContextHeader(ctx)
		req := transhttp.ContextRequest(ctx)
		header.Set(common.HEADER_MSP_ENV_ID, req.Header.Get(common.HEADER_MSP_ENV_ID))
		header.Set(common.HEADER_MSP_AK_ID, req.Header.Get(common.HEADER_MSP_AK_ID))
		header.Set(common.HEADER_MSP_AK_SECRET, req.Header.Get(common.HEADER_MSP_AK_SECRET))
		if data, ok := entity.(*jaegerpb.PostSpansRequest); ok {
			ctx = common.WithSpans(ctx, data.Spans)
		}
		return next(ctx, entity)
	}
}

func ThriftDecoder(r *http.Request, entity interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if _, ok := acceptedThriftFormats[contentType]; !ok {
		return errors.New(fmt.Sprintf("Unsupported content type: %v", html.EscapeString(contentType)))
	}
	if spansRequest, ok := entity.(*jaegerpb.PostSpansRequest); ok {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		tdes := thrift.NewTDeserializer()
		batch := &jaeger.Batch{}
		if err = tdes.Read(r.Context(), batch, body); err != nil {
			return err
		}
		spansRequest.Spans = thrift2Proto(batch)
		extractAuthenticationTags(r, batch.Process.Tags)
	}
	return nil
}

func thrift2Proto(batch *jaeger.Batch) []*tracing.Span {
	if batch == nil || batch.Spans == nil || batch.Process == nil {
		return nil
	}
	blen := len(batch.Spans)
	if blen == 0 {
		return nil
	}
	spans := make([]*tracing.Span, 0)
	for _, tSpan := range batch.Spans {
		span := &tracing.Span{
			TraceID:           extractTraceID(tSpan),
			SpanID:            reflectx.StringToBytes(strconv.FormatInt(tSpan.SpanId, 10)),
			StratTimeUnixNano: uint64(tSpan.StartTime),
			EndTimeUnixNano:   uint64(tSpan.StartTime + tSpan.Duration),
			Name:              tSpan.OperationName,
		}
		if tSpan.ParentSpanId != 0 {
			span.ParentSpanID = reflectx.StringToBytes(strconv.FormatInt(tSpan.SpanId, 10))
		}
		span.Attributes = make(map[string]string)
		span.Attributes[common.TAG_SERVICE_ID] = batch.Process.ServiceName
		span.Attributes[common.TAG_SERVICE_NAME] = batch.Process.ServiceName
		extractAttributes(span, batch.Process.Tags)
		extractAttributes(span, tSpan.Tags)
		spans = append(spans, span)
	}
	return spans
}

func extractTraceID(tSpan *jaeger.Span) []byte {
	if tSpan.TraceIdHigh == 0 {
		return reflectx.StringToBytes(fmt.Sprintf("%016x", tSpan.TraceIdLow))
	}
	return reflectx.StringToBytes(fmt.Sprintf("%016x%016x", tSpan.TraceIdHigh, tSpan.TraceIdLow))
}

func extractAuthenticationTags(r *http.Request, tags []*jaeger.Tag) {
	if tags != nil {
		for _, tag := range tags {
			if tag.Key == common.TAG_MSP_ENV_ID || tag.Key == JAEGER_MSP_ENV_ID {
				// If the headers does not have x-msp-env-id, use msp.env.id contained in tags
				if val := r.Header.Get(common.HEADER_MSP_ENV_ID); val == "" {
					r.Header.Set(common.HEADER_MSP_ENV_ID, extractTagValue(tag))
				}
			}
			if tag.Key == common.TAG_MSP_AK_ID || tag.Key == JAEGER_MSP_AK_ID {
				if val := r.Header.Get(common.HEADER_MSP_AK_ID); val == "" {
					r.Header.Set(common.HEADER_MSP_AK_ID, extractTagValue(tag))
				}
			}
			if tag.Key == common.TAG_MSP_AK_SECRET || tag.Key == JAEGER_MSP_AK_SECRET {
				if val := r.Header.Get(common.HEADER_MSP_AK_SECRET); val == "" {
					r.Header.Set(common.HEADER_MSP_AK_SECRET, extractTagValue(tag))
				}
			}
		}
	}
}

func extractAttributes(span *tracing.Span, tags []*jaeger.Tag) {
	if tags != nil {
		for _, tag := range tags {
			span.Attributes[tag.Key] = extractTagValue(tag)
		}
	}
}

func extractTagValue(tag *jaeger.Tag) string {
	if tag.IsSetVStr() {
		return tag.GetVStr()
	}
	if tag.IsSetVBinary() {
		return reflectx.BytesToString(tag.GetVBinary())
	}
	if tag.IsSetVBool() {
		return strconv.FormatBool(tag.GetVBool())
	}
	if tag.IsSetVDouble() {
		return strconv.FormatFloat(tag.GetVDouble(), 'E', -1, 64)
	}
	if tag.IsSetVLong() {
		return strconv.FormatInt(tag.GetVLong(), 10)
	}
	return tag.GetVStr()
}
