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
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
)

func TestGeneratedFunctions(t *testing.T) {
	// Initialize context with sync.Map
	ctx := context.Background()
	ctx = InitCtxMapIfNeed(ctx)

	// Test RequestID functions
	testRequestID := "test-request-123"
	PutRequestID(ctx, testRequestID)

	if id, ok := GetRequestID(ctx); !ok {
		t.Fatal("GetRequestID should return true when value exists")
	} else if id != testRequestID {
		t.Errorf("Expected %q, got %q", testRequestID, id)
	}

	if id := MustGetRequestID(ctx); id != testRequestID {
		t.Errorf("MustGetRequestID: Expected %q, got %q", testRequestID, id)
	}

	// Test IsStream functions
	PutIsStream(ctx, true)

	if stream, ok := GetIsStream(ctx); !ok {
		t.Fatal("GetIsStream should return true when value exists")
	} else if !stream {
		t.Error("Expected true, got false")
	}

	// Test Logger functions (should NOT have MustGet generated due to custom implementation)
	logger := logrusx.New()
	PutLogger(ctx, logger)

	if l, ok := GetLogger(ctx); !ok {
		t.Fatal("GetLogger should return true when value exists")
	} else if l != logger {
		t.Error("Retrieved logger doesn't match stored logger")
	}

}

func TestGeneratedFunctionsPanic(t *testing.T) {
	// Test that MustGet functions panic when value not found
	ctx := InitCtxMapIfNeed(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRequestID should panic when value not found")
		}
	}()

	MustGetRequestID(ctx)
}

