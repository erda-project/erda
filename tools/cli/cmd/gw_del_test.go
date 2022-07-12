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

package cmd_test

import (
	"testing"

	"github.com/erda-project/erda/tools/cli/cmd"
)

func TestGwDelPreCheck(t *testing.T) {
	var (
		input     = "erda-cloud.invalid-endpoints.json"
		output    string
		cluster   = "erda-cloud"
		kongAdmin = "https://kong-gateway.erda.cloud/"
	)
	if err := cmd.GwDelPreCheck("", &output, cluster, kongAdmin, true); err == nil {
		t.Fatal("err should not be nil")
	}
	if err := cmd.GwDelPreCheck(input, &output, "", kongAdmin, true); err == nil {
		t.Fatal("err should not be nil")
	}
	if err := cmd.GwDelPreCheck(input, &output, cluster, "", true); err == nil {
		t.Fatal("err should not be nil")
	}
	if err := cmd.GwDelPreCheck(input, &output, cluster, kongAdmin, true); err == nil {
		t.Fatal("err should not be nil")
	}
	if err := cmd.GwDelPreCheck(input, &output, cluster, kongAdmin, false); err != nil {
		t.Fatal(err)
	}
}
