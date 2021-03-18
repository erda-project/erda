package apistructs

// 指标集
type Metrics struct {
	Metric []Metric `json:"metrics"`
}

// 指标详情
type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}
