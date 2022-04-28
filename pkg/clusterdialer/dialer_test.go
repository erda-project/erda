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

package clusterdialer

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/discover"
)

const queryIPAddr = "127.0.0.1:18751"

func TestQueryClusterDialerIP(t *testing.T) {
	targetResIP := "testIP"
	http.HandleFunc("/clusterdialer/ip", func(rw http.ResponseWriter, req *http.Request) {
		res := map[string]interface{}{
			"succeeded": true,
			"IP":        targetResIP,
		}
		data, _ := json.Marshal(res)
		io.WriteString(rw, string(data))
	})
	go http.ListenAndServe(queryIPAddr, nil)

	time.Sleep(1 * time.Second)
	os.Setenv(discover.EnvClusterDialer, queryIPAddr)
	res, ok := queryClusterDialerIP("")
	if !ok {
		t.Error("failed to get cluster dialer ip")
	}

	ip, _ := res.(string)
	if ip != targetResIP {
		t.Errorf("got IP: %s, want: %s", ip, targetResIP)
	}
}
