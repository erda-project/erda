package apistructs

type ScheduleLabelListRequest struct{}
type ScheduleLabelListResponse struct {
	Header
	Data ScheduleLabelListData `json:"data"`
}
type ScheduleLabelListData struct {
	// map-key: label name
	// map-value: is this label a prefix?
	Labels map[string]bool `json:"labels"`
}

type ScheduleLabelSetRequest struct {
	// 对于 dcos 的 tag, 由于只有 key, 则 tag 中的 value 都为空
	Tags        map[string]string `json:"tag"`
	Hosts       []string          `json:"hosts"`
	ClusterName string            `json:"clustername"`
	ClusterType string            `json:"clustertype"`
	SoldierURL  string            `json:"soldierURL"`
}

type ScheduleLabelSetResponse struct {
	Header
}
