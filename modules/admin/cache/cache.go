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

package cache

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/cache"
)

var (
	orgID2Org     *cache.Cache
	orgID2OrgName = "admin-org-id-to-org"
)

func init() {
	var bdl = bundle.New(bundle.WithCoreServices())
	orgID2Org = cache.New(orgID2OrgName, time.Minute*10, func(i interface{}) (interface{}, bool) {
		dto, err := bdl.GetOrg(i.(string))
		if err != nil {
			return nil, false
		}
		return dto, true
	})
}

// GetOrgByOrgID gets the *apistructs.OrgDTO by orgID from the newest cache
func GetOrgByOrgID(orgID string) (*apistructs.OrgDTO, bool) {
	item, ok := orgID2Org.LoadWithUpdate(orgID)
	if !ok {
		return nil, false
	}
	return item.(*apistructs.OrgDTO), true
}
