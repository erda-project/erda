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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

const JacksonExpressPrefix = "$."

// use jackson-path command to parse json
func Jackson(jsonInput, filter string) (interface{}, error) {
	f, err := ioutil.TempFile("", "input")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	f.WriteString(jsonInput)
	jq := fmt.Sprintf(`jackson-path -f '%s' -p '%s' -u`, f.Name(), filter)
	wrappedJQ := exec.Command("/bin/sh", "-c", jq)
	output, err := wrappedJQ.CombinedOutput()
	if err != nil {
		logrus.Errorf("jackson failed, filter: %s, input: %s, err: %v", filter, jsonInput, err)
		return "", fmt.Errorf("jackson failed, filter: %s, err: %v, reason: %s", filter, err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
