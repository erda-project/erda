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

package bundle

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/discover"
)

func TestBundle_GetOrg(t *testing.T) {
	os.Setenv(discover.EnvErdaServer, "mock_addr")
	defer os.Unsetenv(discover.EnvErdaServer)
	b := New(WithErdaServer())
	_, err := b.GetOrg("")
	assert.Error(t, err)
	_, err = b.GetOrg(0)
	assert.Error(t, err)
}

func TestBundle_GetDopOrg(t *testing.T) {
	os.Setenv(discover.EnvDOP, "mock_addr")
	defer os.Unsetenv(discover.EnvDOP)
	b := New(WithErdaServer())
	_, err := b.GetDopOrg("")
	assert.Error(t, err)
	_, err = b.GetDopOrg(0)
	assert.Error(t, err)
}
