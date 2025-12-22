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

import "fmt"

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
	switch itemType {
	case ItemTypeModel:
		return "model"
	case ItemTypeProvider:
		return "provider"
	case ItemTypeClientModelRelation:
		return "clientModelRelation"
	case ItemTypeClient:
		return "client"
	case ItemTypeClientToken:
		return "clientToken"
	case ItemTypeTemplate:
		return "template"
	case ItemTypePolicyGroup:
		return "policyGroup"
	default:
		return fmt.Sprintf("unknown(%d)", itemType)
	}
}
