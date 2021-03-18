package apistructs

import "time"

type TestPlanMember struct {
	ID         uint64             `json:"id"`
	TestPlanID uint64             `json:"testPlanID"`
	Role       TestPlanMemberRole `json:"role"`
	UserID     string             `json:"userID"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

type TestPlanMemberRole string

var (
	TestPlanMemberRoleOwner   TestPlanMemberRole = "Owner"
	TestPlanMemberRolePartner TestPlanMemberRole = "Partner"
)

func (role TestPlanMemberRole) Valid() bool {
	switch role {
	case TestPlanMemberRoleOwner, TestPlanMemberRolePartner:
		return true
	default:
		return false
	}
}
func (role TestPlanMemberRole) Invalid() bool {
	return !role.Valid()
}
func (role TestPlanMemberRole) IsOwner() bool {
	return role == TestPlanMemberRoleOwner
}
func (role TestPlanMemberRole) IsPartner() bool {
	return role == TestPlanMemberRolePartner
}
