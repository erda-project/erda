package bundle

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestGetCurrentUser(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	b := New(WithCMDB())
	userInfo, err := b.GetCurrentUser("2")
	assert.NoError(t, err)
	spew.Dump(userInfo)
}

func TestListUsers(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	b := New(WithCMDB())
	userInfo, err := b.ListUsers(apistructs.UserListRequest{
		Query:   "",
		UserIDs: []string{"1", "2"},
	})
	assert.NoError(t, err)
	spew.Dump(userInfo)
}
