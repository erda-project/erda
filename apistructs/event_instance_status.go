package apistructs

// InstanceStatusData 是调度器为实例状态变化事件而定义的结构体
type InstanceStatusData struct {
	ClusterName string `json:"clusterName,omitempty"`
	RuntimeName string `json:"runtimeName,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`

	// 事件id
	// k8s 中是 containerID
	// marathon 中是 taskID
	ID string `json:"id,omitempty"`

	// 容器ip
	IP string `json:"ip,omitempty"`

	// 包含Running,Killed,Failed,Healthy,UnHealthy等状态
	InstanceStatus string `json:"instanceStatus,omitempty"`

	// 宿主机ip
	Host string `json:"host,omitempty"`

	// 事件额外描述，可能为空
	Message string `json:"message,omitempty"`

	// 时间戳到纳秒级
	Timestamp int64 `json:"timestamp"`
}
