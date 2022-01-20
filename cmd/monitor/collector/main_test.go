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

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_getConfigFile(t *testing.T) {
	os.Setenv(envCollectorConfigFile, "aaa.yaml")
	assert.Equal(t, "aaa.yaml", getConfigfile())
	os.Unsetenv(envCollectorConfigFile)

	os.Setenv(string(apistructs.DICE_IS_EDGE), "true")
	assert.Equal(t, edgeConfigFile, getConfigfile())
	os.Unsetenv(string(apistructs.DICE_IS_EDGE))
	os.Unsetenv(envCollectorConfigFile)

	assert.Equal(t, centerConfigFile, getConfigfile())
}
