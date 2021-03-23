package apistructs

import (
	"time"
)

type PipelineTaskDTO struct {
	ID         uint64 `json:"id"`
	PipelineID uint64 `json:"pipelineID"`
	StageID    uint64 `json:"stageID"`

	Name   string             `json:"name"`
	OpType string             `json:"opType"`         // get, put, task
	Type   string             `json:"type,omitempty"` // git, buildpack, release, dice ... 当 OpType 为自定义任务时为空
	Status PipelineStatus     `json:"status"`
	Extra  PipelineTaskExtra  `json:"extra"`
	Labels map[string]string  `json:"labels"`
	Result PipelineTaskResult `json:"result"`

	IsSnippet             bool                       `json:"isSnippet"`
	SnippetPipelineID     *uint64                    `json:"snippetPipelineID,omitempty"`
	SnippetPipelineDetail *PipelineTaskSnippetDetail `json:"snippetPipelineDetail,omitempty" xorm:"json"` // 嵌套的流水线详情

	CostTimeSec  int64     `json:"costTimeSec"`  // -1 表示暂无耗时信息, 0 表示确实是0s结束
	QueueTimeSec int64     `json:"queueTimeSec"` // 等待调度的耗时, -1 暂无耗时信息, 0 表示确实是0s结束 TODO 赋值
	TimeBegin    time.Time `json:"timeBegin"`    // 执行开始时间
	TimeEnd      time.Time `json:"timeEnd"`      // 执行结束时间
	TimeCreated  time.Time `json:"timeCreated"`  // 记录创建时间
	TimeUpdated  time.Time `json:"timeUpdated"`  // 记录更新时间
}

type PipelineTaskExtra struct {
	UUID         string `json:"uuid"`
	AllowFailure bool   `json:"allowFailure"`
}

type PipelineTaskResult struct {
	Metadata    Metadata                 `json:"metadata,omitempty"`
	Errors      []ErrorResponse          `json:"errors,omitempty"`
	MachineStat *PipelineTaskMachineStat `json:"machineStat,omitempty"`
}

type PipelineTaskSnippetDetail struct {
	Outputs []PipelineOutputWithValue `json:"outputs"`

	// 直接子任务数，即 snippet pipeline 的任务数，不会递归查询
	// -1 表示未知，具体数据在 reconciler 调度时赋值
	DirectSnippetTasksNum int `json:"directSnippetTasksNum"`
	// 递归子任务数，即该节点下所有子任务数
	// -1 表示未知，具体数据由 aop 上报
	RecursiveSnippetTasksNum int `json:"recursiveSnippetTasksNum"`
}

type PipelineTaskGetResponse struct {
	Header
	Data *PipelineTaskDTO `json:"data"`
}

type PipelineTaskGetBootstrapInfoResponse struct {
	Header
	Data *PipelineTaskGetBootstrapInfoResponseData `json:"data"`
}

type PipelineTaskGetBootstrapInfoResponseData struct {
	Data []byte `json:"data"`
}

type PipelineTaskMachineStat struct {
	Host PipelineTaskMachineHostStat `json:"host,omitempty"`
	Pod  PipelineTaskMachinePodStat  `json:"pod,omitempty"`
	Load PipelineTaskMachineLoadStat `json:"load,omitempty"`
	Mem  PipelineTaskMachineMemStat  `json:"mem,omitempty"`
	Swap PipelineTaskMachineSwapStat `json:"swap,omitempty"`
}
type PipelineTaskMachineHostStat struct {
	HostIP          string `json:"hostIP,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	UptimeSec       uint64 `json:"uptimeSec,omitempty"`
	BootTimeSec     uint64 `json:"bootTimeSec,omitempty"`
	OS              string `json:"os,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platformVersion,omitempty"`
	KernelVersion   string `json:"kernelVersion,omitempty"`
	KernelArch      string `json:"kernelArch,omitempty"`
}
type PipelineTaskMachinePodStat struct {
	PodIP string `json:"podIP,omitempty"`
}
type PipelineTaskMachineLoadStat struct {
	Load1  float64 `json:"load1,omitempty"`
	Load5  float64 `json:"load5,omitempty"`
	Load15 float64 `json:"load15,omitempty"`
}
type PipelineTaskMachineMemStat struct { // all byte
	Total       uint64  `json:"total,omitempty"`
	Available   uint64  `json:"available,omitempty"`
	Used        uint64  `json:"used,omitempty"`
	Free        uint64  `json:"free,omitempty"`
	UsedPercent float64 `json:"usedPercent,omitempty"`
	Buffers     uint64  `json:"buffers,omitempty"`
	Cached      uint64  `json:"cached,omitempty"`
}
type PipelineTaskMachineSwapStat struct { // all byte
	Total       uint64  `json:"total,omitempty"`
	Used        uint64  `json:"used,omitempty"`
	Free        uint64  `json:"free,omitempty"`
	UsedPercent float64 `json:"usedPercent,omitempty"`
}

type PipelineTaskLoopOptions struct {
	TaskLoop       *PipelineTaskLoop `json:"taskLoop,omitempty"`       // task 指定的 loop 配置
	SpecYmlLoop    *PipelineTaskLoop `json:"specYmlLoop,omitempty"`    // action spec.yml 里指定的 loop 配置
	CalculatedLoop *PipelineTaskLoop `json:"calculatedLoop,omitempty"` // 计算出来的 loop 配置
	LoopedTimes    uint64            `json:"loopedTimes,omitempty"`    // 已循环次数
}

type PipelineTaskLoop struct {
	Break    string        `json:"break" yaml:"break"`
	Strategy *LoopStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty"`
}

func (l *PipelineTaskLoop) Duplicate() *PipelineTaskLoop {
	if l == nil {
		return nil
	}
	d := PipelineTaskLoop{
		Break: l.Break,
	}
	if l.Strategy != nil {
		d.Strategy = &LoopStrategy{
			MaxTimes:        l.Strategy.MaxTimes,
			DeclineRatio:    l.Strategy.DeclineRatio,
			DeclineLimitSec: l.Strategy.DeclineLimitSec,
			IntervalSec:     l.Strategy.IntervalSec,
		}
	}
	return &d
}

var PipelineTaskDefaultLoopStrategy = LoopStrategy{
	MaxTimes:        10, // 默认最多重试 10 ci
	DeclineRatio:    2,  // 默认衰退速率为 2
	DeclineLimitSec: 60, // 默认衰退最大值为 60s
	IntervalSec:     2,  // 默认时间间隔为 5s
}

/**
desc: xxx
priority:
  enable: true
  v1:
    - queue: org-1
      concurrency: 100
      priority: 10
    - queue: project-1
      concurrency: 10
      priority: 20
    - queue: app-i
      concurrency: 1
      priority: 30
*/
type PipelineTaskPriority struct {
	Enable bool                         `json:"enable" yaml:"enable"`
	V1     []PipelineTaskPriorityV1Item `json:"v1" yaml:"v1"`
}

type PipelineTaskPriorityV1Item struct {
	Queue       string `json:"queue" yaml:"queue"`
	Concurrency int64  `json:"concurrency" yaml:"concurrency"`
	Priority    int64  `json:"priority" yaml:"priority"`
}
