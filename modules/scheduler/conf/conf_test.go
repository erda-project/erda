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

package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalEnv(t *testing.T) {
	os.Setenv("DEBUG", "true")
	os.Setenv("LISTEN_ADDR", "*:9091")

	debug := os.Getenv("DEBUG")
	listenAddr := os.Getenv("LISTEN_ADDR")
	assert.Equal(t, "true", debug)
	assert.Equal(t, "*:9091", listenAddr)
}
