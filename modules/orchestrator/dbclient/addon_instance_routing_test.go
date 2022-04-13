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

package dbclient_test

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func TestAddonInstanceRoutingList_GetByName(t *testing.T) {
	var name = "dspo-mysql"
	var list = []dbclient.AddonInstanceRouting{
		{Name: name, Category: "database"},
		{Name: name, Category: apistructs.CUSTOM_TYPE_CUSTOM},
		{Name: name, Category: apistructs.CUSTOM_TYPE_CLOUD},
	}
	l := dbclient.AddonInstanceRoutingList(list)
	item, ok := l.GetByName(name)
	if !ok {
		t.Errorf("not ok, name: %s", name)
	}
	if item.Name != name {
		t.Errorf("name error, expected: %s, actual: %s", name, item.Name)
	}
	if item.Category != apistructs.CUSTOM_TYPE_CUSTOM {
		t.Errorf("category error, expected: %s, actual: %s", apistructs.CUSTOM_TYPE_CUSTOM, item.Category)
	}
}

func TestAddonInstanceRoutingList_GetByTag(t *testing.T) {
	var (
		name = "dspo-mysql"
		tag  = "basic"
	)
	var list = []dbclient.AddonInstanceRouting{
		{Name: name, Category: "database", Tag: tag},
		{Name: name, Category: apistructs.CUSTOM_TYPE_CUSTOM, Tag: tag},
	}
	l := dbclient.AddonInstanceRoutingList(list)
	item, ok := l.GetByTag(tag)
	if !ok {
		t.Errorf("not ok, name: %s", tag)
	}
	if item.Tag != tag {
		t.Errorf("name error, expected: %s, actual: %s", tag, item.Tag)
	}
	if item.Category != apistructs.CUSTOM_TYPE_CUSTOM {
		t.Errorf("category error, expected: %s, actual: %s", apistructs.CUSTOM_TYPE_CUSTOM, item.Category)
	}
}
