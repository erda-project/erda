// Copyright (c) 2022 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dto_test

import (
	"sort"
	"testing"

	"github.com/erda-project/erda/modules/hepa/gateway/dto"
)

func TestSortByTypeList_Less(t *testing.T) {
	var list = dto.SortByTypeList{
		dto.RuntimeDomain{DomainType: dto.EDT_CUSTOM, Domain: "app-custom-1.terminus.io"},
		dto.RuntimeDomain{DomainType: dto.EDT_PACKAGE, Domain: "app-package-1.terminus.io"},
		dto.RuntimeDomain{DomainType: dto.EDT_CUSTOM, Domain: "app-custom3-.terminus.io"},
		dto.RuntimeDomain{DomainType: dto.EDT_DEFAULT, Domain: "app-default.terminus.io"},
		dto.RuntimeDomain{DomainType: dto.EDT_PACKAGE, Domain: "app-package-2.terminus.io"},
		dto.RuntimeDomain{DomainType: dto.EDT_CUSTOM, Domain: "app-custom-2.terminus.io"},
	}
	sort.Sort(list)
	for i := range list {
		t.Logf("[%d] DomainType: %s, Domain: %s", i, list[i].DomainType, list[i].Domain)
	}
}
