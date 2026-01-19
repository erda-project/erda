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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemType_String(t *testing.T) {
	tests := []struct {
		itemType ItemType
		expected string
	}{
		{ItemTypeModel, "model"},
		{ItemTypeProvider, "provider"},
		{ItemTypeClientModelRelation, "clientModelRelation"},
		{ItemTypeClient, "client"},
		{ItemTypeClientToken, "clientToken"},
		{ItemTypeTemplate, "template"},
		{ItemTypePolicyGroup, "policyGroup"},
		{ItemType(999), "unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.itemType.String())
		})
	}
}

func TestIsValidItemType(t *testing.T) {
	validTypes := []string{
		"model",
		"provider",
		"clientModelRelation",
		"client",
		"clientToken",
		"template",
		"policyGroup",
	}

	for _, typeName := range validTypes {
		t.Run(typeName, func(t *testing.T) {
			assert.True(t, IsValidItemType(typeName))
		})
	}

	invalidTypes := []string{
		"unknown",
		"Model", // Case sensitive now
		"PROVIDER",
		"",
	}

	for _, typeName := range invalidTypes {
		t.Run("invalid_"+typeName, func(t *testing.T) {
			assert.False(t, IsValidItemType(typeName))
		})
	}
}

func TestAllItemTypes(t *testing.T) {
	allTypes := AllItemTypes()
	assert.Len(t, allTypes, 7) // Update this count if new types are added

	// Ensure all returned types are valid and have string representations
	for _, itemType := range allTypes {
		assert.NotEmpty(t, itemType.String())
		assert.NotContains(t, itemType.String(), "unknown")
	}
}

func TestItemTypeStrToType(t *testing.T) {
	// Verify bidirectional mapping consistency
	for itemType, name := range ItemTypeToString {
		mappedType, ok := ItemTypeStrToType[name]
		assert.True(t, ok)
		assert.Equal(t, itemType, mappedType)
	}
}
