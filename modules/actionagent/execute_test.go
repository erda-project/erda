package actionagent

//import (
//	"bytes"
//	"encoding/base64"
//	"encoding/json"
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//)
//
//func TestExecute(t *testing.T) {
//	arg := NewAgentArgForPull(10000990, 3154)
//
//	os.Setenv(CONTEXTDIR, "/.pipeline/container/context")
//	os.Setenv(WORKDIR, "/.pipeline/container/context/custom")
//	os.Setenv(METAFILE, "/.pipeline/container/metadata/custom/metadata")
//	os.Setenv("DICE_OPENAPI_URL", "http://openapi.dev.terminus.io")
//	os.Setenv("DICE_OPENAPI_TOKEN", "111")
//	os.Setenv("XXX_PUBLIC_URL", "xxx public url")
//	os.Setenv("IS_EDGE_CLUSTER", "false")
//	os.Setenv("DICE_IS_EDGE", "false")
//
//	b, err := json.Marshal(arg)
//	require.NoError(t, err)
//
//	reqStr := base64.StdEncoding.EncodeToString(b)
//
//	agent := &Agent{}
//	agent.Execute(bytes.NewBufferString(reqStr))
//	fmt.Println("IsEdgeCluster:", agent.EasyUse.IsEdgeCluster)
//	agent.Callback()
//	agent.Exit()
//}
