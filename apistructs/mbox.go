package apistructs

import "time"

type MBoxStatus string

const (
	MBoxReadStatus   MBoxStatus = "read"
	MBoxUnReadStatus MBoxStatus = "unread"
)

// CreateMBoxRequest 创建通知项请求
type CreateMBoxRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	OrgID   int64    `json:"orgId"`
	UserIDs []string `json:"userIds"`
	Label   string   `json:"label"`
}

// CreateMBoxResponse 创建通知项响应
type CreateMBoxResponse struct {
	Header
}

// MBox 站内信结构
type MBox struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Label     string     `json:"label"`
	Status    MBoxStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	ReadAt    *time.Time `json:"readAt"`
}

// QueryMBoxRequest 查询通知发送记录请求
type QueryMBoxRequest struct {
	PageNo   int64      `query:"pageNo"`
	PageSize int64      `query:"pageSize"`
	Label    string     `query:"label"`
	Status   MBoxStatus `query:"status"`
	OrgID    int64      `json:"-"`
	UserID   string     `json:"-"`
}

// QueryMBoxData 站内信记录结构
type QueryMBoxData struct {
	List  []*MBox `json:"list"`
	Total int     `json:"total"`
}

// QueryMBoxResponse 查询通知历史纪录响应
type QueryMBoxResponse struct {
	Header
	Data QueryMBoxData `json:"data"`
}

// QueryMBoxStats 查询站内信统计信息
type QueryMBoxStatsResponse struct {
	Header
	Data QueryMBoxStatsData `json:"data"`
}

type QueryMBoxStatsData struct {
	UnreadCount int `json:"unreadCount"`
}

// SetMBoxReadStatusRequest 标记站内信已读请求
type SetMBoxReadStatusRequest struct {
	OrgID  int64   `json:"-"`
	IDs    []int64 `json:"ids"`
	UserID string  `json:"-"`
}

// SetMBoxReadStatusResponse 批量标记站内信已读响应
type SetMBoxReadStatusResponse struct {
	Header
}
