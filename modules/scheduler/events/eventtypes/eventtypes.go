package eventtypes

type StatusEvent struct {
	Type string `json:"type"`
	// ID由{runtimeName}.{serviceName}.{dockerID}生成
	ID      string `json:"id,omitempty"`
	IP      string `json:"ip,omitempty"`
	Status  string `json:"status"`
	TaskId  string `json:"taskId,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	Host    string `json:"host,omitempty"`
	Message string `json:"message,omitempty"`
}
