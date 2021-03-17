package apistructs

import "time"

// AutoTestSpaceStatus 测试空间状态
type AutoTestSpaceStatus string

func (a AutoTestSpace) IsOpen() bool {
	return a.Status == TestSpaceOpen
}

var (
	// TestSpaceCopying 复制中
	TestSpaceCopying AutoTestSpaceStatus = "copying"
	// TestSpaceLocked 被（复制）锁定
	TestSpaceLocked AutoTestSpaceStatus = "locked"
	// TestSpaceOpen open
	TestSpaceOpen AutoTestSpaceStatus = "open"
	// TestSpaceFailed （复制）失败
	TestSpaceFailed AutoTestSpaceStatus = "failed"
)

// AutoTestSpace 测试空间
type AutoTestSpace struct {
	ID          uint64              `json:"id"`
	Name        string              `json:"name"`
	ProjectID   int64               `json:"projectId"`
	Description string              `json:"description"`
	CreatorID   string              `json:"creatorId"`
	UpdaterID   string              `json:"updaterId"`
	Status      AutoTestSpaceStatus `json:"status"`
	// 被复制的源测试空间
	SourceSpaceID *uint64 `json:"sourceSpaceId,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updatedAt"`
	// DeletedAt 删除时间
	DeletedAt *time.Time `json:"deletedAt"`
}

// AutoTestSpaceCopy 测试空间复制
type AutoTestSpaceCopy struct {
	Name      string `json:"name"`
	SourceID  uint64 `json:"sourceId"`
	ProjectID int64  `json:"projectId"`
}

// AutoTestSpaceCreateRequest 测试空间创建请求
type AutoTestSpaceCreateRequest struct {
	Name          string  `json:"name"`
	ProjectID     int64   `json:"projectId"`
	Description   string  `json:"description"`
	SourceSpaceID *uint64 `json:"sourceSpaceId"`

	IdentityInfo
}

// AutoTestSpaceCreateResponse 测试空间创建响应
type AutoTestSpaceResponse struct {
	Header
	Data *AutoTestSpace `json:"data"`
}

// AutoTestSpaceListResponse 获取测试空间列表响应
type AutoTestSpaceListResponse struct {
	Header
	Data *AutoTestSpaceList `json:"data"`
}

// AutoTestSpaceList 获取测试空间列表
type AutoTestSpaceList struct {
	List  []AutoTestSpace `json:"list"`
	Total int             `json:"total"`
}
