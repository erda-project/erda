package apistructs

// TODO 搜索重构后，此文件可删除
// 搜索类型
const (
	HostSearchType      = "host"
	ContainerSearchType = "container"
	ServiceSearchType   = "service"
	ComponentSearchType = "component"
	AddonSearchType     = "addon"
)

// Resource 查找到的资源
type Resource struct {
	Type     string      `json:"type"`
	Resource interface{} `json:"resource"`
}

// ServiceResource dice上的资源
type ServiceResource struct {
	Name             string                               `json:"name"`
	ProjectUsage     *ProjectUsageFetchResponseData       `json:"project_usage"`
	ApplicationUsage []*ApplicationUsageFetchResponseData `json:"application_usage"`
	RuntimeUsage     []*RuntimeUsageFetchResponseData     `json:"runtime_usage"`
	ServiceUsage     []*ServiceUsageFetchResponseData     `json:"service_usage"`
	Resource         []*ContainerFetchResponseData        `json:"resource"`
}

// ProjectCache 大项目资源缓存
type ProjectCache struct {
	Usage       *ProjectUsageFetchResponseData
	Application map[string]interface{}
	Runtime     map[string]interface{}
	Services    map[string]interface{}
	Resource    []*ContainerFetchResponseData
}

// ExtraUsage 额外资源占用率
type ExtraUsage struct {
	Name     string  `json:"name"`
	Instance int     `json:"instance"`
	Memory   float64 `json:"memory"` // 分配的内存大小单位（MB）
	Disk     float64 `json:"disk"`   // 分配的磁盘大小单位（MB）
	CPU      float64 `json:"cpu"`
}

// ExtraResource 额外的资源，例如中间件，组件
type ExtraResource struct {
	Type     string                       `json:"type"`
	Usage    ExtraUsage                   `json:"usage"`
	Resource []ContainerFetchResponseData `json:"resource"`
}
