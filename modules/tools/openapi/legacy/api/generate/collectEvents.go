package main

import (
        . "github.com/erda-project/erda/apistructs"
)

var Events = [][2]interface{} {
	[2]interface{}{	PipelineEvent{}, `PipelineEvent is k8s-event-like stream event.
`},
	[2]interface{}{	ReleaseEvent{}, `ReleaseEvent Release创建、更新、删除时发送事件
event: release
`},
	[2]interface{}{	RepoCreateMrEvent{}, ``},
	[2]interface{}{	ComponentEvent{}, ``},
	[2]interface{}{	IssueEvent{}, `IssueEvent
`},
	[2]interface{}{	InstanceStatusEvent{}, `InstanceStatusEvent  事件，展示实例的状态变化
event: instances-status
`},
	[2]interface{}{	ClusterEvent{}, `ClusterEvent 创建和修改集群时触发的事件
event: cluster
`},
	[2]interface{}{	GittarPushPayloadEvent{}, ``},
	[2]interface{}{	ExtensionPushEvent{}, `ExtensionPushEvent 扩展更新事件
event: addon_extension_push |action_extension_push
action: create delete update
`},
	[2]interface{}{	GroupNotifyEvent{}, ``},
	[2]interface{}{	PipelineTaskRuntimeEvent{}, `PipelineTaskRuntimeEvent 流水线触发部署时 runtimeID 更新产生的事件
event: pipeline_task_runtime
action: update
`},
	[2]interface{}{	RepoTagEvent{}, `RepoBranchEvent 分支事件
`},
	[2]interface{}{	ApprovalStatusChangedEvent{}, `ApprovalStatusChangedEvent 审批流状态变更事件
`},
	[2]interface{}{	DeleteEvent{}, `DeleteEvent Gittar的删除事件
`},
	[2]interface{}{	PipelineTaskEvent{}, `PipelineTaskEvent 流水线任务状态变化时发送的事件
event: pipeline_task
action: status 见 internal/pipeline/spec/pipeline_status.go#Status
`},
	[2]interface{}{	RepoBranchEvent{}, `RepoBranchEvent 分支事件
`},
	[2]interface{}{	GittarPushEvent{}, `GittarPushEvent POST /callback/gittar eventbox回调的gittar事件结构体
`},
	[2]interface{}{	PipelineInstanceEvent{}, `PipelineInstanceEvent 流水线状态变化时发送的事件
event: pipeline
action: status 见 internal/pipeline/spec/pipeline_status.go#Status
`},
}