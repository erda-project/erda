package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_GetProject(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.marathon.l4lb.thisdcos.directory:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	pj, err := b.GetProject(1)
//	require.NoError(t, err)
//	require.True(t, 1 == pj.ID)
//}
//
//func TestBundle_ListProject(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "10.99.186.34:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	req := apistructs.ProjectListRequest{}
//	req.OrgID = 1
//	req.Name = "dice-dev"
//	req.PageNo = 1
//	req.PageSize = 10
//	pj, err := b.ListProject("2", req)
//	require.NoError(t, err)
//	require.True(t, pj.List[0].Name == "dice-dev")
//}
//
//func TestBundle_GetProjectNS(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	defer func() {
//		os.Unsetenv("CMDB_ADDR")
//	}()
//	b := New(WithCMDB())
//	info, err := b.GetProjectNamespaceInfo(1)
//	require.NoError(t, err)
//	fmt.Println(info)
//}
