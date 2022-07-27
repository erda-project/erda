// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"io"
	"os"
	"strconv"
)

var (
	Namespace = "default"
	NodeName  = os.Getenv("NODE_NAME")
	PodName   = os.Getenv("POD_NAME")
	HostIp    = os.Getenv("HOST_IP")
	PodIp     = os.Getenv("POD_IP")

	ErdaWorkspace = os.Getenv("ERDA_WORKSPACE")
	ErdaCluster   = os.Getenv("ERDA_CLUSTER")
	ErdaOrg       = os.Getenv("ERDA_ORG")
	ErdaProject   = os.Getenv("ERDA_PROJECT")
	ErdaApp       = os.Getenv("ERDA_APP")
)

func init() {
	if v := os.Getenv("NAMESPACE"); v != "" {
		Namespace = v
	}
}

func WriteString(w io.Writer, s string) (n int, err error) {
	return io.WriteString(w, s)
}
func WriteByte(w io.Writer, b byte) (n int, err error) {
	return w.Write([]byte{b})
}
func WriteInt(w io.Writer, i int) (n int, err error) {
	return WriteString(w, strconv.Itoa(i))
}
