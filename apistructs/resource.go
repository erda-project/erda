package apistructs

// /api/scheduler/resource
// method: get
// 查询调度资源
type SchedulerResourceFecthRequest struct {
	Cluster  string            `json:"cluster"`
	Resource SchedulerResource `json:"resource"`
	Attr     Attribute         `json:"attribute"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type SchedulerResource struct {
	CPU  float64 `json:"cpus"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`
}

// Attribute dice_tags like & unlike
type Attribute struct {
	Likes   []string `json:"like"`
	UnLikes []string `json:"unlike"`
}
