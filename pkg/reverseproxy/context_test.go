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

package reverseproxy_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

func TestNewContext(t *testing.T) {
	ctx := reverseproxy.NewContext(map[any]any{
		reverseproxy.MutexCtxKey{}:     "mutex",
		reverseproxy.ProvidersCtxKey{}: "providers",
	})
	mutex, ok := ctx.Value(reverseproxy.MutexCtxKey{}).(string)
	if !ok {
		t.Fatal("failed to retrieve mutex")
	}
	if mutex != "mutex" {
		t.Fatal("mutex is error")
	}
	t.Logf("mutex: %s", mutex)

	providers, ok := ctx.Value(reverseproxy.ProvidersCtxKey{}).(string)
	if !ok {
		t.Fatal("failed to retrieve providers")
	}
	if providers != "providers" {
		t.Fatal("providers is error")
	}
	t.Logf("providers: %s", providers)

	reverseproxy.WithValue(ctx, reverseproxy.FiltersCtxKey{}, "filters")
	filters, ok := ctx.Value(reverseproxy.FiltersCtxKey{}).(string)
	if !ok {
		t.Fatal("failed to retrieve filters")
	}
	if filters != "filters" {
		t.Fatal("filters is error")
	}
	t.Logf("filters: %s", filters)
}
