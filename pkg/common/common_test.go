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

package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findMainEntranceFileName(t *testing.T) {
	mainFileName, found := findMainEntranceFileName()
	assert.False(t, found)
	assert.Empty(t, mainFileName)
}

func Test_loadRootEnvFile(t *testing.T) {
	assert.NoError(t, os.Chdir("testdata"))
	assert.Empty(t, os.Getenv("A"))
	loadRootEnvFile()
	assert.Equal(t, os.Getenv("A"), "B")
}
