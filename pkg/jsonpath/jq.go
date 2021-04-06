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

package jsonpath

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func JQ(jsonInput, filter string) (interface{}, error) {
	f, err := ioutil.TempFile("", "input")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	f.WriteString(jsonInput)
	filter = strings.ReplaceAll(filter, `"`, `\"`)
	filter = filter + " | select(.!=null) | tojson"
	jq := fmt.Sprintf(`cat %s | jq -c -j "%s"`, f.Name(), filter)
	wrappedJQ := exec.Command("/bin/sh", "-c", jq)
	output, err := wrappedJQ.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("jq failed, filter: %s, err: %v, reason: %s", filter, err, string(output))
	}
	if len(output) > 0 {
		var o interface{}
		if err := json.Unmarshal(output, &o); err != nil {
			return nil, err
		}
		return o, nil
	}
	return nil, nil
}
