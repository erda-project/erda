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
	"testing"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy/retry/model_retry"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestFilterRetryExcludedInstances(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	modelretry.AddExcludedModelID(ctx, "m-1")

	in := []*policygroup.RoutingModelInstance{
		{ModelWithProvider: &cachehelpers.ModelWithProvider{Model: &modelpb.Model{Id: "m-1"}}},
		{ModelWithProvider: &cachehelpers.ModelWithProvider{Model: &modelpb.Model{Id: "m-2"}}},
	}
	out := filterRetryExcludedInstances(ctx, in)
	if len(out) != 1 {
		t.Fatalf("expected 1 instance after filtering, got %d", len(out))
	}
	if out[0].ModelWithProvider.Id != "m-2" {
		t.Fatalf("expected m-2 kept, got %s", out[0].ModelWithProvider.Id)
	}
}
