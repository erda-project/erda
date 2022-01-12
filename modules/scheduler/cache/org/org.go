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

package org

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/cache"
)

var orgs *cache.Cache

func init() {
	bdl := bundle.New(bundle.WithCoreServices())
	orgs = cache.New("scheduler-orgs", time.Minute, func(i interface{}) (*cache.Item, bool) {
		orgDTO, err := bdl.GetOrg(i.(string))
		if err != nil {
			return nil, false
		}
		return &cache.Item{
			Object: orgDTO,
		}, true
	})
}

func Get(orgID string) (*apistructs.OrgDTO, bool) {
	item, ok := orgs.LoadWithUpdate(orgID)
	if !ok {
		return nil, false
	}
	return item.Object.(*apistructs.OrgDTO), true
}
