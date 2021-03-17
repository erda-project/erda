package apistructs

import (
	"time"

	"github.com/erda-project/erda/pkg/i18n"
)

type NotifyChannel string

// CreateNotifyRequest 创建通知请求
type CreateNotifyRequest struct {
	Name          string         `json:"name"`
	ScopeType     string         `json:"scopeType"`
	ScopeID       string         `json:"scopeId"`
	Enabled       bool           `json:"enabled"`
	Channels      string         `json:"channels"`
	NotifyGroupID int64          `json:"notifyGroupId"`
	NotifyItemIDs []int64        `json:"notifyItemIds"`
	WithGroup     bool           `json:"withGroup"`
	GroupTargets  []NotifyTarget `json:"groupTargets"`
	Label         string         `json:"label"`
	ClusterName   string         `json:"clusterName"`
	NotifySources []NotifySource `json:"notifySources"`
	WorkSpace     string         `json:"workspace"`
	Creator       string         `json:"-"`
	OrgID         int64          `json:"-"`
}

// CreateNotifyResponse 创建通知响应
type CreateNotifyResponse struct {
	Header
	Data *NotifyDetail `json:"data"`
}

// UpdateNotifyRequest 更新通知请求
type UpdateNotifyRequest struct {
	ID            int64          `json:"id"`
	Channels      string         `json:"channels"`
	NotifyGroupID int64          `json:"notifyGroupId"`
	NotifyItemIDs []int64        `json:"notifyItemIds"`
	NotifySources []NotifySource `json:"notifySources"`
	WithGroup     bool           `json:"withGroup"`
	GroupTargets  []NotifyTarget `json:"groupTargets"`
	GroupName     string         `json:"-"`
	OrgID         int64          `json:"-"`
}

// UpdateNotifyResponse 更新通知响应
type UpdateNotifyResponse struct {
	Header
	Data *NotifyDetail `json:"data"`
}

// DeleteNotifyResponse 删除通知响应
type DeleteNotifyResponse struct {
	Header
	Data *NotifyDetail `json:"data"`
}

// QueryNotifyRequest 查询通知列表请求
type QueryNotifyRequest struct {
	PageNo      int64  `query:"pageNo"`
	PageSize    int64  `query:"pageSize"`
	GroupDetail bool   `query:"groupDetail"`
	ScopeType   string `query:"scopeType"`
	ScopeID     string `query:"scopeId"`
	Label       string `query:"label"`
	ClusterName string `query:"clusterName"`
	OrgID       int64  `json:"-"`
}

// QueryNotifyResponse 查询通知列表响应
type QueryNotifyResponse struct {
	Header
	Data QueryNotifyData `json:"data"`
}

// QueryNotifyData 通知列表数据结构
type QueryNotifyData struct {
	List  []*NotifyDetail `json:"list"`
	Total int             `json:"total"`
}

// Notify 通知
type Notify struct {
	Name      string    `json:"name"`
	ScopeType string    `json:"scopeType"`
	ScopeID   string    `json:"scopeID"`
	Channels  string    `json:"channels"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NotifyDetail 通知详情
type NotifyDetail struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	ScopeType     string          `json:"scopeType"`
	ScopeID       string          `json:"scopeId"`
	Channels      string          `json:"channels"`
	NotifyItems   []*NotifyItem   `json:"notifyItems"`
	NotifyGroup   *NotifyGroup    `json:"notifyGroup"`
	NotifySources []*NotifySource `json:"notifySources"`
	Enabled       bool            `json:"enabled"`
	Creator       string          `json:"creator"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// QuerySourceNotifyResponse 查询通知列表响应
type QuerySourceNotifyResponse struct {
	Header
	Data []*NotifyDetail `json:"data"`
}

// EnableNotifyResponse 启用通知响应
type EnableNotifyResponse struct {
	Header
	Data *NotifyDetail `json:"data"`
}

// DisableNotifyResponse 禁用通知响应
type DisableNotifyResponse struct {
	Header
	Data *NotifyDetail `json:"data"`
}

// FuzzyQueryNotifiesBySourceRequest 模糊查询通知请求
type FuzzyQueryNotifiesBySourceRequest struct {
	SourceType string
	OrgID      int64
	Locale     *i18n.LocaleResource
	Label      string

	// 查询条件
	PageNo      int64
	PageSize    int64
	ClusterName string
	SourceName  string
	NotifyName  string
	ItemName    string
	Channel     string
}
