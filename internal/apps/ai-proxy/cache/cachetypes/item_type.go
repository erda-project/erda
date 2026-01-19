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
	"fmt"
)

// ItemType defines the type of cached items
type ItemType int

const (
	ItemTypeModel ItemType = iota
	ItemTypeProvider
	ItemTypeClientModelRelation
	ItemTypeClient
	ItemTypeClientToken
	ItemTypeTemplate
	ItemTypePolicyGroup
)

func (itemType ItemType) String() string {
	if s, ok := ItemTypeToString[itemType]; ok {
		return s
	}
	return fmt.Sprintf("unknown(%d)", itemType)
}

// ItemTypeToString maps ItemType to string names (camelCase)
var ItemTypeToString = map[ItemType]string{
	ItemTypeModel:               "model",
	ItemTypeProvider:            "provider",
	ItemTypeClientModelRelation: "clientModelRelation",
	ItemTypeClient:              "client",
	ItemTypeClientToken:         "clientToken",
	ItemTypeTemplate:            "template",
	ItemTypePolicyGroup:         "policyGroup",
}

var ItemTypeStrToType map[string]ItemType = func() map[string]ItemType {
	m := make(map[string]ItemType)
	for itemType, name := range ItemTypeToString {
		m[name] = itemType
	}
	return m
}()

// AllItemTypes returns all available item types
func AllItemTypes() []ItemType {
	return []ItemType{
		ItemTypeModel,
		ItemTypeProvider,
		ItemTypeClientModelRelation,
		ItemTypeClient,
		ItemTypeClientToken,
		ItemTypeTemplate,
		ItemTypePolicyGroup,
	}
}

func IsValidItemType(typeName string) bool {
	_, ok := ItemTypeStrToType[typeName]
	return ok
}
