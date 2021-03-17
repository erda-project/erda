package apistructs

type ScopeType string
type PublisherType string

const (
	// SysScope 系统范围
	SysScope ScopeType = "sys"
	// OrgScope 企业范围
	OrgScope ScopeType = "org"
	// ProjectScope 项目范围
	ProjectScope ScopeType = "project"
	// AppScope 应用范围
	AppScope ScopeType = "app"
	// AppScope Publisher范围
	PublisherScope ScopeType = "publisher"
)

// 答疑用户的固定 ID
const SupportID string = "2020"

// Publisher 类型
const (
	// 移动应用
	MobilePublisher PublisherType = "mobile"
)

// 最大scope数量限制
const (
	// MaxOrgNum 最大企业数量限制
	MaxOrgNum uint64 = 5
	// MaxProjectNum 最大项目数量限制
	MaxProjectNum uint64 = 5
	// MaxAppNum 最大应用数量限制
	MaxAppNum uint64 = 5
)

// Scope 范围 (作用域)
type Scope struct {
	// 范围类型
	// 可选值: sys, org, project, app
	Type ScopeType `json:"type"`

	// 范围对应的实例 ID (orgID, projectID, applicationID ...)
	// 比如 type == "org" 时, id 即为 orgID
	ID string `json:"id,omitempty"`
}
