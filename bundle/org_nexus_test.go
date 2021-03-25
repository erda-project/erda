package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestBundle_GetNexusOrgDockerCredentialByImage(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "localhost:9093")
//	defer os.Unsetenv("CMDB_ADDR")
//	bdl := New(WithCMDB())
//	user, err := bdl.GetNexusOrgDockerCredentialByImage(1, "nginx")
//	assert.NoError(t, err)
//	assert.Nil(t, user)
//
//	user, err = bdl.GetNexusOrgDockerCredentialByImage(1, "docker-hosted-org-1-nexus-sys.dev.terminus.io/terminus-dice-dev/test-release-cross-cluster:dockerfile-1593662620685244179")
//	assert.NoError(t, err)
//	assert.NotNil(t, user)
//	spew.Dump(user)
//}
