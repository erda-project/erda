package apistructs

import (
	"time"
)

// ClusterEvent 创建和修改集群时触发的事件
// event: cluster
type ClusterEvent struct {
	EventHeader
	Content ClusterInfo `json:"content"`
}

// InstanceStatusEvent  事件，展示实例的状态变化
// event: instances-status
type InstanceStatusEvent struct {
	EventHeader
	Content InstanceStatusData `json:"content"`
}

// ReleaseEvent Release创建、更新、删除时发送事件
// event: release
type ReleaseEvent struct {
	EventHeader
	Content ReleaseEventData `json:"content"`
}

// PipelineEvent 流水线状态变化时发送的事件
// event: pipeline
// action: status 见 internal/pipeline/spec/pipeline_status.go#Status
type PipelineEvent struct {
	EventHeader
	Content PipelineEventData `json:"content"`
}

type PipelineEventData struct {
	PipelineID      uint64     `json:"pipelineID"`
	Status          string     `json:"status"`
	Branch          string     `json:"branch"`
	Source          string     `json:"source"`          // 来源，dice / qa / bigdata ...
	IsCron          bool       `json:"isCron"`          // 是否是定时触发
	PipelineYmlName string     `json:"pipelineYmlName"` // pipeline yml 文件名
	UserID          string     `json:"userID"`
	InternalClient  string     `json:"internalClient"` // 非用户触发，内部客户端身份
	CostTimeSec     int64      `json:"costTimeSec"`    // 流水线执行耗时
	DiceWorkspace   string     `json:"diceWorkspace"`
	ClusterName     string     `json:"clusterName"`
	TimeBegin       *time.Time `json:"timeBegin"`
	CronExpr        string     `json:"cronExpr"`
}

// PipelineTaskEvent 流水线任务状态变化时发送的事件
// event: pipeline_task
// action: status 见 internal/pipeline/spec/pipeline_status.go#Status
type PipelineTaskEvent struct {
	EventHeader
	Content PipelineTaskEventData `json:"content"`
}

type PipelineTaskEventData struct {
	PipelineTaskID  uint64    `json:"pipelineTaskID"`
	PipelineID      uint64    `json:"pipelineID"`
	ActionType      string    `json:"actionType"` // git, custom ...
	Status          string    `json:"status"`
	ClusterName     string    `json:"clusterName"` // 集群名
	UserID          string    `json:"userID"`
	CreatedAt       time.Time `json:"createdAt"`
	QueueTimeSec    int64     `json:"queueTimeSec"` // 排队耗时
	CostTimeSec     int64     `json:"costTimeSec"`  // 任务执行耗时 (不包含 排队耗时)
	OrgName         string    `json:"orgName"`
	ProjectName     string    `json:"projectName"`
	ApplicationName string    `json:"applicationName"`
	TaskName        string    `json:"taskName"`
	RuntimeID       string    `json:"runtimeID"`
	ReleaseID       string    `json:"releaseID"`
}

// PipelineTaskRuntimeEvent 流水线触发部署时 runtimeID 更新产生的事件
// event: pipeline_task_runtime
// action: update
type PipelineTaskRuntimeEvent struct {
	EventHeader
	Content PipelineTaskRuntimeEventData `json:"content"`
}

type PipelineTaskRuntimeEventData struct {
	ClusterName    string `json:"clusterName"` // 集群名
	PipelineTaskID uint64 `json:"pipelineTaskID"`
	Status         string `json:"status"`
	RuntimeID      string `json:"runtimeID"`
}

type GroupNotifyEvent struct {
	Sender  string                 `json:"sender"`
	Content GroupNotifyContent     `json:"content"`
	Lables  map[string]interface{} `json:"lables"`
}

type GroupNotifyContent struct {
	SourceName            string               `json:"sourceName"`
	SourceType            string               `json:"sourceType"`
	SourceID              string               `json:"sourceId"`
	NotifyName            string               `json:"notifyName"`
	NotifyItemDisplayName string               `json:"notifyItemDisplayName"`
	Channels              []GroupNotifyChannel `json:"channels"`
	OrgID                 int64                `json:"orgId"`
	Label                 string               `json:"label"`
	ClusterName           string               `json:"clusterName"`
	CalledShowNumber      string               `json:"calledShowNumber"`
}

type GroupNotifyChannel struct {
	Name     string            `json:"name"`
	Template string            `json:"template"`
	Type     string            `json:"type"` // 用于mail模式渲染 值为markdown会二次渲染html
	Tag      string            `json:"tag"`  //  用于webhook的附加信息
	Params   map[string]string `json:"params"`
}

// ExtensionPushEvent 扩展更新事件
// event: addon_extension_push |action_extension_push
// action: create delete update
type ExtensionPushEvent struct {
	EventHeader
	Content ExtensionPushEventData `json:"content"`
}

type ExtensionPushEventData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

// ApprovalStatusChangedEvent 审批流状态变更事件
type ApprovalStatusChangedEvent struct {
	EventHeader
	Content ApprovalStatusChangedEventData `json:"content"`
}

// ApprovalStatusChangedEventData 审批流状态变更事件数据
type ApprovalStatusChangedEventData struct {
	ApprovalID     uint64         `json:"approvalID"`
	ApprovalStatus ApprovalStatus `json:"approvalStatus"`
	ApprovalType   ApproveType    `json:"approvalType"`
}

// IssueEvent
type IssueEvent struct {
	EventHeader
	Content IssueEventData `json:"content"`
}

// IssueEventData
type IssueEventData struct {
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	AtUserIDs    string          `json:"atUserIds"`
	StreamType   IssueStreamType `json:"streamType"`
	StreamParams ISTParam        `json:"streamParams"`
}
