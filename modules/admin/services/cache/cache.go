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

package cache

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/cache"
)

var (
	c *Cache

	orgID2Org *cache.Cache
	proID2Org *cache.Cache
)

type Cache struct{}

// GetOrgByOrgID gets the *apistructs.OrgDTO by orgID from the newest cache
func (c *Cache) GetOrgByOrgID(orgID string) (*apistructs.OrgDTO, bool) {
	item, ok := orgID2Org.LoadWithUpdate(orgID)
	if !ok {
		return nil, false
	}
	return item.(*apistructs.OrgDTO), true
}

// GetOrgByProjectID gets the *apistructs.OrgDTO by projectID from the newest cache
func (c *Cache) GetOrgByProjectID(projectID string) (*apistructs.OrgDTO, bool) {
	item, ok := proID2Org.LoadWithUpdate(projectID)
	if !ok {
		return nil, false
	}
	return item.(*apistructs.OrgDTO), true
}
