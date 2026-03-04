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

package health

import (
	"context"
	"sync"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

type PolicyGroupHealthMeta struct {
	FilteredUnhealthyInstanceIDs []string
	ReleasedUnsupportedCount     int
	ReleasedUnsupportedAPITypes  []string

	mu                          sync.Mutex
	filteredUnhealthyInstanceID map[string]struct{}
	releasedUnsupportedAPIType  map[string]struct{}
}

func NewPolicyGroupHealthMeta() *PolicyGroupHealthMeta {
	return &PolicyGroupHealthMeta{
		filteredUnhealthyInstanceID: make(map[string]struct{}),
		releasedUnsupportedAPIType:  make(map[string]struct{}),
	}
}

func ensurePolicyGroupHealthMeta(ctx context.Context) *PolicyGroupHealthMeta {
	metaVal, ok := ctxhelper.GetPolicyGroupHealthMeta(ctx)
	if ok && metaVal != nil {
		if meta, ok := metaVal.(*PolicyGroupHealthMeta); ok && meta != nil {
			return meta
		}
	}
	meta := NewPolicyGroupHealthMeta()
	ctxhelper.PutPolicyGroupHealthMeta(ctx, meta)
	return meta
}

func AppendFilteredUnhealthyInstanceID(ctx context.Context, instanceID string) {
	if ctx == nil || instanceID == "" {
		return
	}
	meta := ensurePolicyGroupHealthMeta(ctx)
	meta.AddFilteredUnhealthyInstanceID(instanceID)
}

func AppendReleasedUnsupportedAPIType(ctx context.Context, apiType string) {
	if ctx == nil || apiType == "" {
		return
	}
	meta := ensurePolicyGroupHealthMeta(ctx)
	meta.AddReleasedUnsupportedAPIType(apiType)
}

func (m *PolicyGroupHealthMeta) AddFilteredUnhealthyInstanceID(instanceID string) {
	if m == nil || instanceID == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.filteredUnhealthyInstanceID[instanceID]; exists {
		return
	}
	m.filteredUnhealthyInstanceID[instanceID] = struct{}{}
	m.FilteredUnhealthyInstanceIDs = append(m.FilteredUnhealthyInstanceIDs, instanceID)
}

func (m *PolicyGroupHealthMeta) AddReleasedUnsupportedAPIType(apiType string) {
	if m == nil || apiType == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReleasedUnsupportedCount++
	if _, exists := m.releasedUnsupportedAPIType[apiType]; exists {
		return
	}
	m.releasedUnsupportedAPIType[apiType] = struct{}{}
	m.ReleasedUnsupportedAPITypes = append(m.ReleasedUnsupportedAPITypes, apiType)
}
