package apistructs

import "time"

// addon base request
type AddonOpsBaseRequest struct {
	ClusterName string `json:"clusterName"`
	ProjectName string `json:"projectName"`
	ProjectID   string `json:"projectID"`
	AddonID     string `json:"addonID"`
	AddonName   string `json:"addonName"`
}

// addon scale request
type AddonScaleRequest struct {
	AddonOpsBaseRequest
	// CPU cpu大小
	CPU float64 `json:"cpu"`
	// Mem 内存大小
	Mem uint64 `json:"mem"`
	// Nodes 节点数量
	Nodes int `json:"nodes"`
}

type AddonScaleResponse Header

// addon project quota check request
type ProjectQuotaCheckRequest AddonScaleRequest

type BaseResource struct {
	CPU float64 `json:"cpu"`
	// Mem 内存大小
	Mem float64 `json:"mem"`
}

type ProjectQuotaCheckResponse struct {
	IsQuotaEnough bool         `json:"isQuotaEnough"`
	Remain        BaseResource `json:"remain"`
	Need          BaseResource `json:"need"`
}

// addon config request
type AddonConfigRequest struct {
	AddonID string `json:"addonID"`
}

type AddonConfigResponse struct {
	Header
	Data *AddonConfigData `json:"data"`
}

type AddonConfigData struct {
	AddonID    string            `json:"addonID"`
	AddonName  string            `json:"addonName"`
	Config     map[string]string `json:"config"`
	CPU        float64           `json:"cpu"`
	Mem        uint64            `json:"mem"`
	Nodes      int               `json:"nodes"`
	CreateTime time.Time         `json:"createTime"`
	UpdateTime time.Time         `json:"updateTime"`
}

// addon config update request
type AddonConfigUpdateRequest struct {
	AddonOpsBaseRequest

	// 更新配置信息，覆盖更新
	Config map[string]string `json:"config"`
}

type AddonConfigUpdateResponse Header

// addon status response
type OpsAddonStatusResponse struct {
	Header
	Data OpsAddonStatusData `json:"data"`
}

// addon status query status
type OpsAddonStatusQueryRequest struct {
	AddonName string `query:"addonName"`
	AddonID   string `query:"addonID"`
}

type OpsAddonStatusData struct {
	Status StatusCode `json:"status"`
}
