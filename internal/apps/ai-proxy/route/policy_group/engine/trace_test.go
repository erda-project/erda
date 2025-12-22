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

package engine

import (
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestBuildRouteTrace_SetsFallbackFromSticky(t *testing.T) {
	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group: &pb.PolicyGroup{
			Name:      "g1",
			StickyKey: "x-request-id",
		},
	}

	trace := buildRouteTrace(req, "v1", true, nil)
	if trace.Sticky == nil || trace.Sticky.Value != "v1" {
		t.Fatalf("expected sticky.value=v1, got %+v", trace.Sticky)
	}
	if trace.Sticky == nil || !trace.Sticky.FallbackFromSticky {
		t.Fatalf("expected fallbackFromSticky=true, got false")
	}
}
