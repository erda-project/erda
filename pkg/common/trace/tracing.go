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

	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/otel/trace"
)

func TracerStartSpan(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	ctx = trace.ContextWithSpan(ctx, span)
	return ctx
}

func TracerFinishSpan(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)
	span.Finish()
}
