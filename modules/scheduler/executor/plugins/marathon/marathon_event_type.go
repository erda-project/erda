package marathon

type MIpas struct {
	IpAddress string `json:"ipAddress"`
}

type MarathonStatusUpdateEvent struct {
	TaskStatus  string  `json:"taskStatus"`
	IpAddresses []MIpas `json:"ipAddresses"`
	AppId       string  `json:"appId"`
	TaskId      string  `json:"taskId"`
	Host        string  `json:"host"`
	Message     string  `json:"message"`
}

type MarathonInstanceHealthChangedEvent struct {
	InstanceId string `json:"instanceId"`
	Healthy    bool   `json:"healthy"`
}
