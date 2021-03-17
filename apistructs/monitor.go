package apistructs

// GetMonitorResponse .
type GetMonitorAlertResponse struct {
	Header
	Data *Alert `json:"data"`
}

// Alert 告警结构体
type Alert struct {
	ID           uint64                 `json:"id"`
	Name         string                 `json:"name"`
	AlertScope   string                 `json:"alertScope"`
	AlertScopeID string                 `json:"alertScopeId"`
	Enable       bool                   `json:"enable"`
	Attributes   map[string]interface{} `json:"attributes"`
	CreateTime   int64                  `json:"createTime"`
	UpdateTime   int64                  `json:"updateTime"`
}

// GetMonitorReportTaskResponse .
type GetMonitorReportTaskResponse struct {
	Header
	Data *ReportTask `json:"data"`
}

// ReportTask
type ReportTask struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	ScopeID     string `json:"scopeId"`
	DashboardId string `json:"dashboardId"`
	Enable      bool   `json:"enable"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
