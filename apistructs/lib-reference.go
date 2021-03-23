package apistructs

import "time"

// LibReference 库引用返回结构
type LibReference struct {
	ID             uint64         `json:"id"`
	AppID          uint64         `json:"appID"`
	LibID          uint64         `json:"libID"`
	LibName        string         `json:"libName"`
	LibDesc        string         `json:"libDesc"`
	ApprovalID     uint64         `json:"approvalID"`
	ApprovalStatus ApprovalStatus `json:"approvalStatus"`
	Creator        string         `json:"creator"`
	CreatedAt      *time.Time     `json:"createdAt"`
	UpdatedAt      *time.Time     `json:"updatedAt"`
}

// LibReferenceCreateRequest 库引用创建请求
type LibReferenceCreateRequest struct {
	AppID   uint64 `json:"appID"`
	AppName string `json:"appName"`
	LibID   uint64 `json:"libID"`
	LibName string `json:"libName"`
	LibDesc string `json:"libDesc"`

	OrgID uint64
	IdentityInfo
}

// LibReferenceCreateResponse 库引用创建响应
type LibReferenceCreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// LibReferenceListRequest 库引用请求
type LibReferenceListRequest struct {
	// +optional
	AppID uint64 `schema:"appID"`
	// +optional
	LibID uint64 `schema:"libID"`
	// +optional
	ApprovalStatus ApprovalStatus `schema:"approvalStatus"`
	// +optional
	PageNo uint64 `schema:"pageNo"`
	// +optional
	PageSize uint64 `schema:"pageSize"`

	IdentityInfo
}

// LibReferenceListResponse 库引用响应
type LibReferenceListResponse struct {
	Header
	UserInfoHeader
	Data LibReferenceListResponseData `json:"data"`
}

// LibReferenceListResponseData 库引用响应数据
type LibReferenceListResponseData struct {
	Total uint64         `json:"total"`
	List  []LibReference `json:"list"`
}

// LibReferenceVersion 库引用版本
type LibReferenceVersion struct {
	LibName string `json:"libName"`
	Version string `json:"version"`
}
