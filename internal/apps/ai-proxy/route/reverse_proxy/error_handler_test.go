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

package reverse_proxy

import (
	"context"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestNoteAttemptCompleted(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	sink := types.New("audit-1", nil)
	ctxhelper.PutAuditSink(ctx, sink)

	NoteAttemptCompleted(ctx)

	got := sink.Snapshot()
	if got["response_at"] == nil {
		t.Fatal("expected response_at to be recorded")
	}
	if got["response_chunk_done_at"] == nil {
		t.Fatal("expected response_chunk_done_at to be recorded")
	}
}
