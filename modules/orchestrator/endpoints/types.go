package endpoints

type RuntimeStatusEventReq struct {
	RuntimeName     string               `json:"runtimeName"`
	IsDeleted       bool                 `json:"isDeleted"`
	EventType       string               `json:"eventType"` // total(全量事件)/increment(增量事件)
	ServiceStatuses []ServiceStatusEvent `json:"serviceStatuses"`
}

type ServiceStatusEvent struct {
	ServiceName      string                `json:"serviceName"`
	Status           string                `json:"serviceStatus"`
	Replica          int                   `json:"replica"`
	InstanceStatuses []InstanceStatusEvent `json:"instanceStatuses"`
}

type InstanceStatusEvent struct {
	TaskId string                 `json:"id"` // TaskId
	IP     string                 `json:"ip"`
	Status string                 `json:"instanceStatus"`
	Stage  string                 `json:"stage"`
	Extra  map[string]interface{} `json:"extra"`
}
