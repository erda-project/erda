package nexus

//import (
//	"encoding/json"
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNexus_ListUsers(t *testing.T) {
//	users, err := n.ListUsers(UserListRequest{
//		UserIDPrefix: "",
//		Source:       "default",
//	})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&users, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_GetUsers(t *testing.T) {
//	user, err := n.GetUser("admin")
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&user, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_ensureUser(t *testing.T) {
//	err := n.ensureUser(UserEnsureRequest{UserCreateRequest: UserCreateRequest{
//		UserID:       "maven-hosted-publisher-1-deployment",
//		FirstName:    "maven-hosted-publisher-1-deployment",
//		LastName:     "maven-hosted-publisher-1-deployment",
//		EmailAddress: "maven-hosted-publisher-1-deployment@terminus.io",
//		Password:     "1234567",
//		Status:       UserStatusActive,
//		Roles:        []RoleID{"nx-admin"},
//	}})
//	assert.NoError(t, err)
//}
//
//func TestNexus_CreateUser(t *testing.T) {
//	err := n.CreateUser(UserCreateRequest{
//		UserID:       "maven-hosted-publisher-1-deployment",
//		FirstName:    "maven-hosted-publisher-1-deployment",
//		LastName:     "maven-hosted-publisher-1-deployment",
//		EmailAddress: "maven-hosted-publisher-1-deployment@terminus.io",
//		Password:     "123456",
//		Status:       UserStatusActive,
//		Roles:        []RoleID{"nx-anonymous"},
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_UpdateUser(t *testing.T) {
//	err := n.UpdateUser(UserUpdateRequest{
//		UserCreateRequest: UserCreateRequest{
//			UserID:       "maven-hosted-publisher-1-deployment",
//			FirstName:    "maven-hosted-publisher-1-deployment",
//			LastName:     "maven-hosted-publisher-1-deployment",
//			EmailAddress: "maven-hosted-publisher-1-deployment@terminus.io",
//			Status:       UserStatusActive,
//			Roles:        []RoleID{"nx-admin"},
//		},
//		Source:        "default",
//		ReadOnly:      true,
//		ExternalRoles: nil,
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_ChangeUserPassword(t *testing.T) {
//	err := n.ChangeUserPassword(UserChangePasswordRequest{
//		UserID:   "maven-hosted-publisher-1-deployment",
//		Password: "1234567",
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_EnsureUser(t *testing.T) {
//	userName := "org-group-user"
//	userID, err := n.EnsureUser(EnsureUserRequest{
//		UserName: userName,
//		Password: "12345678",
//		RepoPrivileges: map[RepositoryFormat]map[string][]PrivilegeAction{
//			RepositoryFormatMaven: {
//				"maven-releases": []PrivilegeAction{PrivilegeActionREAD},
//				"maven-central":  []PrivilegeAction{PrivilegeActionREAD},
//			},
//			//RepositoryFormatNpm: {
//			//	"npm-hosted-publisher-24": []PrivilegeAction{PrivilegeActionREAD, PrivilegeActionBROWSE},
//			//},
//		},
//	})
//	assert.NoError(t, err)
//	assert.Equal(t, userID, UserID(userName))
//}
