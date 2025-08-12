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
)

// MustGetResponseChunkIndex retrieves ResponseChunkIndex from context, returns -1 if not found
// This overrides the generated MustGetResponseChunkIndex to provide special fallback logic
func MustGetResponseChunkIndex(ctx context.Context) int {
	index, ok := GetResponseChunkIndex(ctx)
	if !ok {
		return -1
	}
	return index
}

func IncrementResponseChunkIndex(ctx context.Context) int {
	index, ok := GetResponseChunkIndex(ctx)
	if !ok {
		PutResponseChunkIndex(ctx, 1)
		return 1
	}
	newIndex := index + 1
	PutResponseChunkIndex(ctx, newIndex)
	return newIndex
}

func IsFirstResponseChunk(ctx context.Context) bool {
	index, ok := GetResponseChunkIndex(ctx)
	return !ok || index == 0
}
