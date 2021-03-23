package apistructs

// AppWorkspaceReleasesGetRequest 查询应用某个环境所有可部署的 release 请求
type AppWorkspaceReleasesGetRequest struct {
	AppID     uint64        `schema:"appID,required"`
	Workspace DiceWorkspace `schema:"workspace,required"`
}

// AppWorkspaceReleasesGetResponse 查询应用某个环境所有可部署的 release 响应
type AppWorkspaceReleasesGetResponse struct {
	Header
	Data AppWorkspaceReleasesGetResponseData `json:"data,omitempty"`
}

// AppWorkspaceReleasesGetResponseData map key: branch, map value: paging releases
type AppWorkspaceReleasesGetResponseData map[string]*ReleaseListResponseData
