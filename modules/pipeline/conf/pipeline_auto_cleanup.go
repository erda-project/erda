package conf

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
)

// PipelineAutoCleanupStrategy 自动清理策略配置
type PipelineAutoCleanupStrategy map[apistructs.PipelineSource]PipelineAutoCleanupSourceStrategyItem

// PipelineAutoCleanupSourceStrategyItem 配置 source 级别流水线自动清理策略
type PipelineAutoCleanupSourceStrategyItem struct {
	// -- Finished

	// ArchiveFinished
	// true: 归档
	// false: 直接删除表记录，不做归档
	ArchiveFinished *bool `json:"archiveFinished"`

	// MaxFinishedStoreTime 已完成记录保存的最大时间
	MaxFinishedStoreTime    time.Duration `json:"-"`
	MaxFinishedStoreTimeStr string        `json:"maxFinishedStoreTime"`

	// MinFinishedStoreCount 已完成记录保存至少多少条
	// <-1 使用默认保留条数
	// =-1 不清理
	// >=0 保留 n 条
	MinFinishedStoreCount *int64 `json:"minFinishedStoreCount"`

	// -- Analyzed

	// ArchiveAnalyzed
	// true: 归档
	// false: 直接删除表记录，不做归档
	ArchiveAnalyzed *bool `json:"archiveAnalyzed"`

	// MaxAnalyzedStoreTime 未开始记录保存的最大时间
	MaxAnalyzedStoreTime    time.Duration `json:"-"`
	MaxAnalyzedStoreTimeStr string        `json:"maxAnalyzedStoreTime"`

	// MaxAnalyzedStoreTime 未开始记录保存至少多少条
	// <-1 使用默认保留条数
	// =-1 不清理
	// >=0 保留 n 条
	MinAnalyzedStoreCount *int64 `json:"minAnalyzedStoreCount"`
}

func (item PipelineAutoCleanupSourceStrategyItem) String() string {
	return fmt.Sprintf("MaxFinishedStoreTime: %s, MinFinishedStoreCount: %d, "+"MaxAnalyzedStoreTime: %s, MinAnalyzedStoreCount: %d",
		item.MaxFinishedStoreTime, item.MinFinishedStoreCount, item.MaxAnalyzedStoreTime, item.MinAnalyzedStoreCount)
}

const defaultSource = "default"

// GetSourceStrategy 获取 source 配置
func (s PipelineAutoCleanupStrategy) GetSourceStrategy(source apistructs.PipelineSource) PipelineAutoCleanupSourceStrategyItem {
	config, ok := s[source]
	if !ok {
		return s[defaultSource]
	}
	return config
}

// handlePipelineAutoCleanupStrategy 处理 source 默认值
func handlePipelineAutoCleanupStrategy(cfg *Conf) {
	// 将 storeTimeStr 转换为 storeTime
	for source, item := range cfg.PipelineAutoCleanupStrategy {
		// MaxFinishedStoreTime
		if item.MaxFinishedStoreTimeStr != "" {
			d, err := time.ParseDuration(item.MaxFinishedStoreTimeStr)
			if err != nil {
				panic(err)
			}
			item.MaxFinishedStoreTime = d
		}
		// MaxAnalyzedStoreTime
		if item.MaxAnalyzedStoreTimeStr != "" {
			d, err := time.ParseDuration(item.MaxAnalyzedStoreTimeStr)
			if err != nil {
				panic(err)
			}
			item.MaxAnalyzedStoreTime = d
		}
		cfg.PipelineAutoCleanupStrategy[source] = item
	}

	// 设置默认配置
	defaultSourceCfg := cfg.PipelineAutoCleanupStrategy[defaultSource]
	// 已完成 默认归档
	if defaultSourceCfg.ArchiveFinished == nil {
		defaultSourceCfg.ArchiveFinished = &[]bool{true}[0]
	}
	// 已完成 默认保存 30 天
	if defaultSourceCfg.MaxFinishedStoreTime <= 0 {
		defaultSourceCfg.MaxFinishedStoreTime = time.Hour * 24 * 30
	}
	// 已完成 默认至少保留 100 条
	if defaultSourceCfg.MinFinishedStoreCount == nil || *defaultSourceCfg.MinFinishedStoreCount < -1 {
		defaultSourceCfg.MinFinishedStoreCount = &[]int64{100}[0]
	}
	// 未开始 默认不归档，直接删除
	if defaultSourceCfg.ArchiveAnalyzed == nil {
		defaultSourceCfg.ArchiveAnalyzed = &[]bool{false}[0]
	}
	// 未开始 默认保存 1 天
	if defaultSourceCfg.MaxAnalyzedStoreTime <= 0 {
		defaultSourceCfg.MaxAnalyzedStoreTime = time.Hour * 24
	}
	// 未开始 默认至少保留 10 条
	if defaultSourceCfg.MinAnalyzedStoreCount == nil || *defaultSourceCfg.MinAnalyzedStoreCount < -1 {
		defaultSourceCfg.MinAnalyzedStoreCount = &[]int64{10}[0]
	}
	if cfg.PipelineAutoCleanupStrategy == nil {
		cfg.PipelineAutoCleanupStrategy = make(map[apistructs.PipelineSource]PipelineAutoCleanupSourceStrategyItem)
	}
	cfg.PipelineAutoCleanupStrategy[defaultSource] = defaultSourceCfg

	// 遍历 source，配置未空，则使用默认值
	for source, item := range cfg.PipelineAutoCleanupStrategy {
		if item.ArchiveFinished == nil {
			item.ArchiveFinished = defaultSourceCfg.ArchiveFinished
		}
		if item.MaxFinishedStoreTime <= 0 {
			item.MaxFinishedStoreTime = defaultSourceCfg.MaxFinishedStoreTime
		}
		if item.MinFinishedStoreCount == nil || *item.MinFinishedStoreCount < -1 {
			item.MinFinishedStoreCount = defaultSourceCfg.MinFinishedStoreCount
		}
		if item.ArchiveAnalyzed == nil {
			item.ArchiveAnalyzed = defaultSourceCfg.ArchiveAnalyzed
		}
		if item.MaxAnalyzedStoreTime <= 0 {
			item.MaxAnalyzedStoreTime = defaultSourceCfg.MaxAnalyzedStoreTime
		}
		if item.MinAnalyzedStoreCount == nil || *item.MinAnalyzedStoreCount < -1 {
			item.MinAnalyzedStoreCount = defaultSourceCfg.MinAnalyzedStoreCount
		}
		cfg.PipelineAutoCleanupStrategy[source] = item
	}
}
