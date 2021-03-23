package apistructs

// AutoopOutputLine 自动化运维脚本执行时输出的行内容
type AutoopOutputLine struct {
	Stream string `json:"stream"`
	Node   string `json:"node"`
	Host   string `json:"host"`
	Body   string `json:"body"`
}
