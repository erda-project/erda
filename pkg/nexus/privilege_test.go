package nexus

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNxRepositoryViewPrivileges(t *testing.T) {
	fmt.Println(
		GetNxRepositoryViewPrivileges(
			"npm-hosted-publisher-1-deployment",
			RepositoryFormatMaven,
			PrivilegeActionADD,
			PrivilegeActionEDIT,
			PrivilegeActionBROWSE,
			PrivilegeActionREAD,
		),
	)
}

func TestNexus_ListPrivileges(t *testing.T) {
	users, err := n.ListPrivileges(PrivilegeListRequest{})
	assert.NoError(t, err)
	s, _ := json.MarshalIndent(&users, "", "  ")
	fmt.Println(string(s))
}

func TestNexus_GetPrivilege(t *testing.T) {
	users, err := n.GetPrivilege(PrivilegeGetRequest{
		PrivilegeID: "test-content-selector-privilege-maven-all",
	})
	assert.NoError(t, err)
	s, _ := json.MarshalIndent(&users, "", "  ")
	fmt.Println(string(s))
}

func TestNexus_DeletePrivilege(t *testing.T) {
	err := n.DeletePrivilege(PrivilegeDeleteRequest{
		PrivilegeID: "nx-all",
	})
	assert.NoError(t, err)
}

func TestNexus_CreateRepositoryContentSelectorPrivilege(t *testing.T) {
	privilegeName := "test-content-selector-privilege-npm"

	err := n.CreateRepositoryContentSelectorPrivilege(RepositoryContentSelectorPrivilegeCreateRequest{
		Name:            privilegeName,
		Description:     "all maven repo read",
		Actions:         []PrivilegeAction{PrivilegeActionREAD},
		Format:          RepositoryFormatNpm,
		Repository:      "*",
		ContentSelector: "test-content-selector",
	})
	assert.NoError(t, err)

	privilege, err := n.GetPrivilege(PrivilegeGetRequest{
		PrivilegeID: privilegeName,
	})
	assert.NoError(t, err)
	printJSON(privilege)
}

func TestNexus_UpdateRepositoryContentSelectorPrivilege(t *testing.T) {
	privilegeName := "test-content-selector-privilege-npm"

	err := n.UpdateRepositoryContentSelectorPrivilege(RepositoryContentSelectorPrivilegeUpdateRequest{
		Name:            privilegeName,
		Description:     "all maven repo readsss",
		Actions:         []PrivilegeAction{PrivilegeActionREAD},
		Format:          RepositoryFormatNpm,
		Repository:      "*",
		ContentSelector: "test-content-selector",
	})
	assert.NoError(t, err)

	privilege, err := n.GetPrivilege(PrivilegeGetRequest{
		PrivilegeID: privilegeName,
	})
	assert.NoError(t, err)
	printJSON(privilege)
}
