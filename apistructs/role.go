package apistructs

// /api/roles/<role>
// method: put
// 更改用户角色
type RoleChangeRequest struct {
	// 用户角色
	Role string         `json:"-" path:"role"`
	Body RoleChangeBody `json:"body"`
}

type RoleChangeBody struct {
	// 目标id, 对应的applicationId, projectId, orgId
	TargetId string `json:"targetId"`

	// 目标类型 APPLICATION,PROJECT,ORG
	TargetType string `json:"targetType"`
	UserId     string `json:"userId"`
}

// /api/roles/<role>
// method: put
type RoleChangeResponse struct {
	Header
	Data bool `json:"data"`
}
