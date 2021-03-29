package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func TestBundle_GetApp(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	_, err := b.GetApp(1)
//	require.NoError(t, err)
//}
//
//func TestBundle_GetMyApp(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	apps, err := b.GetMyApps("2", 1)
//	// require.NoError(t, err)
//	assert.NoError(t, err)
//	fmt.Println(apps)
//}
