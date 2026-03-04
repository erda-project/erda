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

package context

import (
	"context"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/require"

	audittypes "github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type fakeAuditSink struct {
	notes map[string]any
}

func (s *fakeAuditSink) Note(key string, val any) {
	s.notes[key] = val
}

func (s *fakeAuditSink) NoteOnce(key string, val any) bool {
	if _, ok := s.notes[key]; ok {
		return false
	}
	s.notes[key] = val
	return true
}

func (s *fakeAuditSink) NoteAppend(key string, val any) {
	if old, ok := s.notes[key]; ok {
		s.notes[key] = []any{old, val}
		return
	}
	s.notes[key] = val
}

func (s *fakeAuditSink) Inc(key string, delta int64) int64 {
	var base int64
	if old, ok := s.notes[key].(int64); ok {
		base = old
	}
	base += delta
	s.notes[key] = base
	return base
}

func (s *fakeAuditSink) Flush(_ context.Context, _ audittypes.Writer) {}

func (s *fakeAuditSink) Snapshot() map[string]any {
	out := make(map[string]any, len(s.notes))
	for k, v := range s.notes {
		out[k] = v
	}
	return out
}

func TestSaveContextToAuditWritesPolicyGroupHealthMeta(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	sink := &fakeAuditSink{notes: map[string]any{}}
	ctxhelper.PutAuditSink(ctx, sink)
	health.AppendFilteredUnhealthyInstanceID(ctx, "i1")
	health.AppendFilteredUnhealthyInstanceID(ctx, "i2")
	health.AppendReleasedUnsupportedAPIType(ctx, "embeddings")

	req := httptest.NewRequest("POST", "http://example.com/v1/chat/completions", nil).WithContext(ctx)
	req.Header.Set(vars.XAIProxySource, "unit-test")
	pr := &httputil.ProxyRequest{
		In:  req,
		Out: req.Clone(ctx),
	}

	err := (&Context{}).saveContextToAudit(pr)
	require.NoError(t, err)

	require.Equal(t, 2, sink.notes["policy_group.health.filtered_unhealthy_count"])
	require.Equal(t, 1, sink.notes["policy_group.health.released_unsupported_count"])
	require.Equal(t, []string{"i1", "i2"}, sink.notes["policy_group.health.filtered_unhealthy_instance_ids"])
	require.Equal(t, []string{"embeddings"}, sink.notes["policy_group.health.released_unsupported_api_types"])
}
