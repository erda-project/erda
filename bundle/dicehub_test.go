package bundle

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestBundle_QueryExtensionVersions(t *testing.T) {
	os.Setenv("DICEHUB_ADDR", "http://127.0.0.1:10000")
	bdl := New(WithDiceHub())

	result, err := bdl.QueryExtensionVersions(apistructs.ExtensionVersionQueryRequest{
		Name: "mysql",
	})
	assert.Nil(t, err)
	t.Log(result)

	os.Unsetenv("DICEHUB_ADDR")
}

func TestBundle_GetExtensionVersion(t *testing.T) {
	os.Setenv("DICEHUB_ADDR", "http://127.0.0.1:10000")
	bdl := New(WithDiceHub())

	result, err := bdl.GetExtensionVersion(apistructs.ExtensionVersionGetRequest{
		Name:    "mysql",
		Version: "0.0.3",
	})
	assert.Nil(t, err)
	t.Log(result)

	os.Unsetenv("DICEHUB_ADDR")
}

func TestBundle_SearchExtensions(t *testing.T) {
	os.Setenv("DICEHUB_ADDR", "dicehub.default.svc.cluster.local:10000")
	bdl := New(WithDiceHub())

	result, err := bdl.SearchExtensions(apistructs.ExtensionSearchRequest{
		Extensions: []string{"git"},
	})
	assert.Nil(t, err)
	t.Log(result)

	os.Unsetenv("DICEHUB_ADDR")
}
