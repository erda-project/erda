// Copyright (c) 2021 Terminus, Inc.
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

func (g TuneGroup) RegisterTuneChainByTypeAndTrigger(pointType TuneType, trigger TuneTrigger) TuneChain {
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
