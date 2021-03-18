package bundle

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBundle_GetOrg(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	defer func() {
		os.Unsetenv("CMDB_ADDR")
	}()
	b := New(WithCMDB())
	org, err := b.GetOrg(1)
	require.NoError(t, err)
	fmt.Println(org.Domain)

}
