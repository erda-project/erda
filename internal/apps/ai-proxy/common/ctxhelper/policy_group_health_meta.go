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

package ctxhelper

import (
	"context"
	"sync"
)

type mapKeyPolicyGroupHealthMeta struct {
	*PolicyGroupHealthMeta
}

type PolicyGroupHealthMeta struct {
	FilteredUnhealthyInstanceIDs []string
	ReleasedUnsupportedCount     int
	ReleasedUnsupportedAPITypes  []string

	mu                          sync.Mutex
	filteredUnhealthyInstanceID map[string]struct{}
	releasedUnsupportedAPIType  map[string]struct{}
}

func ensurePolicyGroupHealthMeta(ctx context.Context) *PolicyGroupHealthMeta {
	meta, ok := getFromMapKeyAs[*PolicyGroupHealthMeta](ctx, mapKeyPolicyGroupHealthMeta{})
	if ok && meta != nil {
		return meta
	}
	meta = &PolicyGroupHealthMeta{
		filteredUnhealthyInstanceID: make(map[string]struct{}),
		releasedUnsupportedAPIType:  make(map[string]struct{}),
	}
	putToMapKey(ctx, mapKeyPolicyGroupHealthMeta{}, meta)
	return meta
}

func GetPolicyGroupHealthMeta(ctx context.Context) (*PolicyGroupHealthMeta, bool) {
	meta, ok := getFromMapKeyAs[*PolicyGroupHealthMeta](ctx, mapKeyPolicyGroupHealthMeta{})
	return meta, ok && meta != nil
}

func AppendFilteredUnhealthyInstanceID(ctx context.Context, instanceID string) {
	if ctx == nil || instanceID == "" {
		return
	}
	meta := ensurePolicyGroupHealthMeta(ctx)
	meta.mu.Lock()
	defer meta.mu.Unlock()
	if _, exists := meta.filteredUnhealthyInstanceID[instanceID]; exists {
		return
	}
	meta.filteredUnhealthyInstanceID[instanceID] = struct{}{}
	meta.FilteredUnhealthyInstanceIDs = append(meta.FilteredUnhealthyInstanceIDs, instanceID)
}

func AppendReleasedUnsupportedAPIType(ctx context.Context, apiType string) {
	if ctx == nil || apiType == "" {
		return
	}
	meta := ensurePolicyGroupHealthMeta(ctx)
	meta.mu.Lock()
	defer meta.mu.Unlock()
	meta.ReleasedUnsupportedCount++
	if _, exists := meta.releasedUnsupportedAPIType[apiType]; exists {
		return
	}
	meta.releasedUnsupportedAPIType[apiType] = struct{}{}
	meta.ReleasedUnsupportedAPITypes = append(meta.ReleasedUnsupportedAPITypes, apiType)
}
