package apistructs

// CICDPipelineListRequest /api/cicds 获取 pipeline 列表
type CICDPipelineListRequest struct {
	Branches string `schema:"branches"`
	Sources  string `schema:"sources"`
	YmlNames string `schema:"ymlNames"`
	Statuses string `schema:"statuses"`
	AppID    uint64 `schema:"appID"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

// CICDPipelineYmlListRequest /api/cicds/actions/pipelineYmls 获取 pipeline yml列表
type CICDPipelineYmlListRequest struct {
	AppID  int64  `schema:"appID"`
	Branch string `schema:"branch"`
}

// CICDPipelineYmlListResponse
type CICDPipelineYmlListResponse struct {
	Data []string `json:"data"`
}
