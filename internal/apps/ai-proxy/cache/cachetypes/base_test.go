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

package cachetypes

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
)

func TestSmartCloneProtoMessage(t *testing.T) {
	original := &modelpb.Model{
		ApiKey: "secret",
		Metadata: &metadatapb.Metadata{
			Secret: map[string]*structpb.Value{
				"token": structpb.NewStringValue("value"),
			},
		},
	}

	cloned := smartClone(original).(*modelpb.Model)

	if original == cloned {
		t.Fatalf("expected different pointer after clone")
	}
	assert.Equal(t, original.ApiKey, cloned.ApiKey)

	cloned.ApiKey = "changed"
	assert.Equal(t, "secret", original.ApiKey)
}

func TestSmartCloneSliceOfProtoMessages(t *testing.T) {
	original := []*modelpb.Model{
		{
			ApiKey: "secret-1",
		},
		nil,
	}

	cloned := smartClone(original).([]*modelpb.Model)

	assert.Equal(t, len(original), len(cloned))
	assert.Nil(t, cloned[1])
	if original[0] == cloned[0] {
		t.Fatalf("expected slice element to be cloned")
	}

	cloned[0].ApiKey = "changed"
	assert.Equal(t, "secret-1", original[0].ApiKey)
}

func TestSmartCloneFallbackDeepCopy(t *testing.T) {
	type sample struct {
		Values map[string]int
	}

	original := sample{
		Values: map[string]int{"foo": 1},
	}

	cloned := smartClone(original).(sample)

	if fmt.Sprintf("%p", original.Values) == fmt.Sprintf("%p", cloned.Values) {
		t.Fatalf("expected map to be cloned")
	}
	cloned.Values["foo"] = 2
	assert.Equal(t, 1, original.Values["foo"])
}

func TestSmartCloneNilSlice(t *testing.T) {
	var original []*modelpb.Model

	cloned := smartClone(original)
	// After the fix, nil slice should become empty slice
	assert.NotNil(t, cloned, "nil slice should become non-nil after clone")

	// Type assertion should succeed
	typedResult, ok := cloned.([]*modelpb.Model)
	assert.True(t, ok, "type assertion should succeed")
	assert.NotNil(t, typedResult)
	assert.Equal(t, 0, len(typedResult))
}

// TestEmptySliceTypeAssertion verifies that empty slices (non-nil) can be
// successfully type-asserted, which is crucial for cache operations.
// This ensures that QueryFromDB implementations return empty slices instead of nil.
func TestEmptySliceTypeAssertion(t *testing.T) {
	// Empty slice (non-nil) - this should succeed in type assertion
	emptySlice := []*modelpb.Model{}
	var anyVal any = emptySlice

	// Type assertion should NOT panic
	result, ok := anyVal.([]*modelpb.Model)
	assert.True(t, ok, "type assertion for empty slice should succeed")
	assert.NotNil(t, result, "empty slice should not be nil after type assertion")
	assert.Equal(t, 0, len(result), "empty slice should have length 0")

	// Also verify that ranging over empty slice works
	count := 0
	for range result {
		count++
	}
	assert.Equal(t, 0, count, "ranging over empty slice should work")
}

// TestNilSliceTypeAssertionPanic demonstrates why nil slices cause issues.
// When resp.List is nil, type assertion .([]T) will panic.
func TestNilSliceTypeAssertionPanic(t *testing.T) {
	// nil value - this would fail type assertion without the fix
	var nilSlice []*modelpb.Model = nil
	var anyValNil any = nilSlice

	// This assertion will succeed but result will be nil
	result, ok := anyValNil.([]*modelpb.Model)
	assert.True(t, ok, "type assertion for nil slice actually succeeds")
	assert.Nil(t, result, "nil slice remains nil after type assertion")

	// But if the any value is interface nil (not typed nil), it WILL fail:
	var untyped any = nil
	result2, ok2 := untyped.([]*modelpb.Model)
	assert.False(t, ok2, "type assertion for untyped nil should fail")
	assert.Nil(t, result2)
}

// TestEnsureNonNilSlice verifies the framework-level nil slice handling.
func TestEnsureNonNilSlice(t *testing.T) {
	t.Run("nil slice becomes empty slice", func(t *testing.T) {
		var nilSlice []*modelpb.Model = nil
		result := ensureNonNilSlice(nilSlice)

		// Should be non-nil after processing
		assert.NotNil(t, result, "nil slice should become non-nil")

		// Type assertion should succeed
		typedResult, ok := result.([]*modelpb.Model)
		assert.True(t, ok, "type assertion should succeed")
		assert.NotNil(t, typedResult, "typed result should be non-nil")
		assert.Equal(t, 0, len(typedResult), "should be empty slice")
	})

	t.Run("non-nil slice remains unchanged", func(t *testing.T) {
		original := []*modelpb.Model{{Id: "test"}}
		result := ensureNonNilSlice(original)

		typedResult, ok := result.([]*modelpb.Model)
		assert.True(t, ok)
		assert.Equal(t, 1, len(typedResult))
		assert.Equal(t, "test", typedResult[0].Id)
	})

	t.Run("empty slice remains empty", func(t *testing.T) {
		original := []*modelpb.Model{}
		result := ensureNonNilSlice(original)

		typedResult, ok := result.([]*modelpb.Model)
		assert.True(t, ok)
		assert.NotNil(t, typedResult)
		assert.Equal(t, 0, len(typedResult))
	})

	t.Run("nil interface returns nil", func(t *testing.T) {
		var nilInterface any = nil
		result := ensureNonNilSlice(nilInterface)
		assert.Nil(t, result)
	})
}

// mockQueryFromDB simulates IQueryFromDB interface implementation
type mockQueryFromDB struct{}

func (m *mockQueryFromDB) QueryFromDB(ctx context.Context) (uint64, any, error) {
	// Simulate database returning nil list (empty table)
	var list []*modelpb.Model = nil
	return 0, list, nil
}

func (m *mockQueryFromDB) GetIDValue(item any) (string, error) {
	model := item.(*modelpb.Model)
	return model.Id, nil
}

// TestEnsureNonNilSlice_WithMockQueryFromDB verifies the complete flow
// when QueryFromDB returns a nil slice (simulating empty database table).
func TestEnsureNonNilSlice_WithMockQueryFromDB(t *testing.T) {
	mock := &mockQueryFromDB{}

	// Call QueryFromDB which returns nil slice
	_, data, err := mock.QueryFromDB(context.Background())
	assert.NoError(t, err)

	// Before ensureNonNilSlice: data holds a typed nil
	t.Logf("Before ensureNonNilSlice:")
	t.Logf("  data == nil: %v", data == nil)
	t.Logf("  reflect.ValueOf(data).Kind(): %v", reflect.ValueOf(data).Kind())
	t.Logf("  reflect.ValueOf(data).IsNil(): %v", reflect.ValueOf(data).IsNil())

	// Key assertion: typed nil is NOT equal to interface nil
	assert.False(t, data == nil, "typed nil slice should NOT be equal to interface nil")

	// Apply ensureNonNilSlice (this is what the framework does)
	result := ensureNonNilSlice(data)

	// After ensureNonNilSlice: result is a non-nil empty slice
	t.Logf("After ensureNonNilSlice:")
	t.Logf("  result == nil: %v", result == nil)

	assert.NotNil(t, result, "result should be non-nil after ensureNonNilSlice")

	// Type assertion should succeed
	models, ok := result.([]*modelpb.Model)
	assert.True(t, ok, "type assertion should succeed")
	assert.NotNil(t, models, "models should be non-nil empty slice")
	assert.Equal(t, 0, len(models), "models should be empty")

	// Verify we can safely range over it
	count := 0
	for range models {
		count++
	}
	assert.Equal(t, 0, count, "ranging over empty slice should work")
}
