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

package meta

import (
	"context"
	"time"

	"github.com/appscode/go/strings"
)

type Interface interface {
	GetMeta(ctx context.Context, scope, scopeId string, names ...string) []*MetricMeta
	WaitAndGetTables(ctx context.Context) map[MetricUniq]*MetricMeta
	Reload() chan error
}

func (p *provider) GetMeta(ctx context.Context, scope, scopeId string, names ...string) []*MetricMeta {
	metas, ok := p.Meta.Load().(map[MetricUniq]*MetricMeta)
	if !ok {
		return nil
	}

	var result []*MetricMeta

	for k, v := range metas {
		if len(names) > 1 {
			if indexOf(names, k.MetricGroup) == -1 {
				continue
			}
		}
		if k.Scope != scope {
			continue
		}
		if !strings.IsEmpty(&scopeId) {
			if k.ScopeId != scopeId {
				continue
			}
		}
		result = append(result, v)
	}
	return result
}

func indexOf(arr []string, target string) int {
	for i, x := range arr {
		if x == target {
			return i
		}
	}
	return -1
}

func (p *provider) WaitAndGetTables(ctx context.Context) map[MetricUniq]*MetricMeta {
	for {
		metes, ok := p.Meta.Load().(map[MetricUniq]*MetricMeta)

		if ok && len(metes) > 0 {
			return metes
		}

		// wait for the index to complete loading
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *provider) Reload() chan error {
	ch := make(chan error, 1)
	p.reloadCh <- ch
	return ch
}
