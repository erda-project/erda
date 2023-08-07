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

package project_report

import (
	"strconv"

	"github.com/patrickmn/go-cache"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/legacy/model"
)

type orgCache struct {
	*cache.Cache
}

func (o *orgCache) Get(key uint64) *orgpb.Org {
	org, ok := o.Cache.Get(strconv.FormatUint(key, 10))
	if !ok {
		return nil
	}
	return org.(*orgpb.Org)
}

func (o *orgCache) Set(key uint64, orgDto *orgpb.Org) {
	o.Cache.Set(strconv.FormatUint(key, 10), orgDto, cache.NoExpiration)
}

type projectCache struct {
	*cache.Cache
}

func (p *projectCache) Get(key int64) *model.Project {
	projectDto, ok := p.Cache.Get(strconv.FormatInt(key, 10))
	if !ok {
		return nil
	}
	return projectDto.(*model.Project)
}

func (p *projectCache) Set(key int64, projectDto *model.Project) {
	p.Cache.Set(strconv.FormatInt(key, 10), projectDto, cache.NoExpiration)
}

type iterationCache struct {
	*cache.Cache
}

func (i *iterationCache) Get(key uint64) *IterationInfo {
	iter, ok := i.Cache.Get(strconv.FormatUint(key, 10))
	if !ok {
		return nil
	}
	return iter.(*IterationInfo)
}

func (i *iterationCache) Set(key uint64, iter *IterationInfo) {
	i.Cache.Set(strconv.FormatUint(key, 10), iter, cache.NoExpiration)
}

func (i *iterationCache) Iterate(fn func(k string, v interface{}) error) {
	for k, v := range i.Cache.Items() {
		if err := fn(k, v.Object); err != nil {
			return
		}
	}
}
