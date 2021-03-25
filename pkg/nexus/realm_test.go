package nexus

//import (
//	"encoding/json"
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNexus_ListActiveRealms(t *testing.T) {
//	users, err := n.ListActiveRealms(RealmListActiveRequest{})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&users, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_SetActiveRealms(t *testing.T) {
//	err := n.SetActiveRealms(RealmSetActivesRequest{
//		ActiveRealms: []RealmID{
//			NpmBearerTokenRealm.ID,
//		},
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_ListAvailableRealms(t *testing.T) {
//	realms, err := n.ListAvailableRealms(RealmListAvailableRequest{})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&realms, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_EnsureAddRealms(t *testing.T) {
//	assert.NoError(t, n.EnsureAddRealms(RealmEnsureAddRequest{Realms: []RealmID{DockerTokenRealm.ID}}))
//}
