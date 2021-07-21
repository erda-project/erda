package common

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	Default = "Default"

	NodeStatusReady int = iota
	NodeStatusError
	NodeStatusFreeze
)

type (
	SteveStatusEnum string
	UsageStatusEnum string
)

type SteveStatus struct {
	Value      SteveStatusEnum `json:"value,omitempty"`
	RenderType string          `json:"render_type"`
	Status     SteveStatusEnum `json:"status"`
}

var (
	/*
		node phase
	*/
	NodeSuccess SteveStatusEnum = "success"
	NodeDefault SteveStatusEnum = "default"
	NodeFreeze  SteveStatusEnum = "freeze"
	NodeError   SteveStatusEnum = "error"

	NodeSuccessCN SteveStatusEnum = "正常"
	NodeDefaultCN SteveStatusEnum = "默认"
	NodeFreezeCN  SteveStatusEnum = "冻结"
	NodeErrorCN   SteveStatusEnum = "节点错误"

	// NodeReady means kubelet is healthy and ready to accept pods.
	NodeReady SteveStatusEnum = "Ready"
	// NodeMemoryPressure means the kubelet is under pressure due to insufficient available memory.
	NodeMemoryPressure SteveStatusEnum = "MemoryPressure"
	// NodeDiskPressure means the kubelet is under pressure due to insufficient available disk.
	NodeDiskPressure SteveStatusEnum = "DiskPressure"
	// NodePIDPressure means the kubelet is under pressure due to insufficient available PID.
	NodePIDPressure SteveStatusEnum = "PIDPressure"
	// NodeNetworkUnavailable means that network for the node is not correctly configured.
	NodeNetworkUnavailable SteveStatusEnum = "NetworkUnavailable"

	/*
		pod status
	*/
	PodRunning   = SteveStatusEnum(v1.PodRunning)
	PodPending   = SteveStatusEnum(v1.PodPending)
	PodSuccessed = SteveStatusEnum(v1.PodSucceeded)
	PodFailed    = SteveStatusEnum(v1.PodFailed)
	PodUnknown   = SteveStatusEnum(v1.PodUnknown)

	PodRunningCN   SteveStatusEnum = "运行"
	PodPendingCN   SteveStatusEnum = "预备"
	PodSuccessedCN SteveStatusEnum = "退出成功"
	PodFailedCN    SteveStatusEnum = "退出错误"
	PodUnknownCN   SteveStatusEnum = "未知"

	/*
		resource usage status
	*/
	ResourceDefault UsageStatusEnum = "default"
	ResourceSafe    UsageStatusEnum = "safe"
	ResourceWarning UsageStatusEnum = "warning"
	ResourceDanger  UsageStatusEnum = "danger"

	ResourceDefaultCN UsageStatusEnum = "默认"
	ResourceSafeCN    UsageStatusEnum = "安全"
	ResourceWarningCN UsageStatusEnum = "警告"
	ResourceDangerCN  UsageStatusEnum = "危险"

	/*
		resource usage status
	*/
	//WorkflowDefault UsageStatusEnum = "default"
	//ResourceSafe    UsageStatusEnum = "safe"
	//ResourceWarning UsageStatusEnum = "warning"
	//ResourceDanger  UsageStatusEnum = "danger"

)
var (
	NodeNotFoundErr           = errors.New("node not found")
	PodNotFoundErr            = errors.New("pod not found")
	OperationsEmptyErr        = errors.New("operation is empty")
	NodeStatusEmptyErr        = errors.New("node status not found")
	ProtocolComponentEmptyErr = errors.New("component is nil or property empty")
)
var nodeStatusMap = map[int][]SteveStatusEnum{
	NodeStatusReady:  {NodeSuccess, NodeSuccessCN},
	NodeStatusFreeze: {NodeFreeze, NodeFreezeCN},
	NodeStatusError:  {NodeError, NodeErrorCN},
}

var podStatusMap = map[SteveStatusEnum][]SteveStatusEnum{
	PodRunning:   {PodRunning, PodRunningCN},
	PodPending:   {PodPending, PodPendingCN},
	PodSuccessed: {PodSuccessed, PodSuccessedCN},
	PodFailed:    {PodFailed, PodFailedCN},
	PodUnknown:   {PodUnknown, PodUnknownCN},
}

type State struct {
	IsFirstFilter   bool                   `json:"is_first_filter"`
	PageNo          int                    `json:"page_no"`
	PageSize        int                    `json:"page_size"`
	Total           int                    `json:"total"`
	Query           map[string]interface{} `json:"query"`
	SelectedRowKeys []string               `json:"selected_row_keys"`
	Start           time.Time              `json:"start"`
	End             time.Time              `json:"end"`
	Name            string                 `json:"name"`
	ClusterName     string                 `json:"cluster_name"`
	Namespace       string                 `json:"namespace"`
	SortColumnName  string                 `json:"sorter"`
}

type ChartDataItem struct {
	Value float64 `json:"value"`
	Time  int64   `json:"time"`
}
