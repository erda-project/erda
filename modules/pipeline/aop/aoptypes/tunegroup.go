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
