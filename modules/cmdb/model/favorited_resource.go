package model

// FavoritedResource 收藏的资源
type FavoritedResource struct {
	BaseModel
	Target   string // 被收藏的资源类型: app/project, etc
	TargetID uint64
	UserID   string
}
