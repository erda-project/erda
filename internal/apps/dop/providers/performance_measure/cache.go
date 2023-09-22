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

package performance_measure

import (
	"fmt"
	"strconv"

	"github.com/patrickmn/go-cache"
)

type personalPerformanceCache struct {
	*cache.Cache
}

func (p *personalPerformanceCache) Get(key uint64) *PersonalPerformanceInfo {
	info, ok := p.Cache.Get(strconv.FormatUint(key, 10))
	if !ok {
		return nil
	}
	return info.(*PersonalPerformanceInfo)
}

func (p *personalPerformanceCache) Set(key uint64, info *PersonalPerformanceInfo) {
	p.Cache.Set(strconv.FormatUint(key, 10), info, cache.NoExpiration)
}

func (p *personalPerformanceCache) Iterate(fn func(k string, v interface{}) error) {
	for k, v := range p.Cache.Items() {
		if err := fn(k, v.Object); err != nil {
			return
		}
	}
}

type propertyCache struct {
	*cache.Cache
}

func genWontfixStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("wontfix_state_ids_%d", projectID)
}

func genOpenStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("open_state_ids_%d", projectID)
}

func genWorkingStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("working_state_ids_%d", projectID)
}

func genClosedStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("closed_state_ids_%d", projectID)
}

func genDoneStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("doen_state_ids_%d", projectID)
}

func genReopenStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("reopen_state_ids_%d", projectID)
}

func genResolvedStateIDSKey(projectID uint64) string {
	return fmt.Sprintf("resolved_state_ids_%d", projectID)
}

func (p *propertyCache) SetWontfixStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genWontfixStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GetWonfixStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genWontfixStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetOpenStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genOpenStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GetOpenStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genOpenStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetWorkingStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genWorkingStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GetWorkingStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genWorkingStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetClosedStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genClosedStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GetClosedStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genClosedStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetDoneStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genDoneStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GetDoneStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genDoneStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetReopenStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genReopenStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GeReopenStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genReopenStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}

func (p *propertyCache) SetResolvedStateIDS(projectID uint64, stateIDS []uint64) {
	p.Cache.Set(genResolvedStateIDSKey(projectID), stateIDS, cache.NoExpiration)
}

func (p *propertyCache) GeResolvedStateIDS(projectID uint64) []uint64 {
	stateIDS, ok := p.Cache.Get(genResolvedStateIDSKey(projectID))
	if !ok {
		return nil
	}
	return stateIDS.([]uint64)
}
