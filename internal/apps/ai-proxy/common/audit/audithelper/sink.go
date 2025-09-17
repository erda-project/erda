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

package audithelper

import (
	"context"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/notes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func Note(ctx context.Context, k string, v any) {
	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok || sink == nil {
		return
	}
	sink.Note(k, v)
}

func NoteAppend(ctx context.Context, k string, v any) {
	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok || sink == nil {
		return
	}
	sink.NoteAppend(k, v)
}

func Flush(ctx context.Context) {
	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok || sink == nil {
		return
	}
	sinkWriter := notes.NewDBWriter(ctxhelper.MustGetDBClient(ctx))
	sink.Flush(ctx, sinkWriter)
}
