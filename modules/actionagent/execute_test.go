// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
