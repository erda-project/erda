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

package aoptypes

// TuneGroup 保存所有类型不同触发时机下的调用链
type TuneGroup map[TuneType]map[TuneTrigger]TuneChain

// GetTuneChainByTypeAndTrigger 根据 类型 和 触发时机 返回 调用链
func (g TuneGroup) GetTuneChainByTypeAndTrigger(pointType TuneType, trigger TuneTrigger) TuneChain {
	if len(g) == 0 {
		return nil
	}
	// type
	chains, ok := g[pointType]
	if !ok || len(chains) == 0 {
		return nil
	}
	// trigger
	return chains[trigger]
}
