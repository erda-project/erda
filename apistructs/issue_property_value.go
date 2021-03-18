package apistructs

// 创建字段枚举值请求
type IssuePropertyValueCreateRequest struct {
	Value      string `json:"value"`
	PropertyID int64  `json:"propertyID"`
	IdentityInfo
}

// 更新字段枚举值请求
type IssuePropertyValueUpdateRequest struct {
	PropertyID int64  `json:"propertyID"` // 字段ID
	ID         int64  `json:"id"`
	Value      string `json:"value"`
	IdentityInfo
}

// 删除字段枚举值请求
type IssuePropertyValueDeleteRequest struct {
	PropertyValueID int64 `json:"propertyID"` // 字段ID
	IdentityInfo
}

// 查询字段枚举值请求
type IssuePropertyValueGetRequest struct {
	PropertyID int64 `json:"propertyID"` // 企业
	IdentityInfo
}
