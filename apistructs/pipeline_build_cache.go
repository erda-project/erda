package apistructs

type BuildCacheImageReportRequest struct {
	Action      string `json:"action"`
	Name        string `json:"name"`
	ClusterName string `json:"clusterName"`
}

type BuildCacheImageReportResponse struct {
	Header
}
