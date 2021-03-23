package apistructs

type MemeberLabelName string

const (
	// LabelNameOutsource 外包人员
	LabelNameOutsource MemeberLabelName = "Outsource"
	// LabelNamePartner 合作伙伴
	LabelNamePartner MemeberLabelName = "Partner"
)

// ResourceKey member关联字段的key
type ExtraResourceKey string

const (
	// LabelResourceKey 成员的标签
	LabelResourceKey ExtraResourceKey = "label"
	// RoleResourceKey 成员的角色
	RoleResourceKey ExtraResourceKey = "role"
)

func (rk ExtraResourceKey) String() string {
	return string(rk)
}

// MemberLabelListResponse 查询成员标签列表 GET /api/members/actions/list-labels
type MemberLabelListResponse struct {
	Header
	Data MemberLabelList `json:"data"`
}

// MemberLabelList 成员标签列表
type MemberLabelList struct {
	// 角色标签
	List []MemberLabelInfo `json:"list"`
}

// MemberLabelInfo 成员标签
type MemberLabelInfo struct {
	Label MemeberLabelName `json:"label"`
	Name  string           `json:"name"`
}
