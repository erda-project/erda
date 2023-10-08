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

package csrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getHostPort(t *testing.T) {
	assert.Equal(t, "localhost:8080", getHostPort("localhost:8080", ""))
	assert.Equal(t, "localhost:8080", getHostPort("localhost:8080", "http"))
	assert.Equal(t, "localhost:8080", getHostPort("localhost:8080", "https"))
	assert.Equal(t, "localhost:443", getHostPort("localhost", "https"))
	assert.Equal(t, "localhost:80", getHostPort("localhost", "http"))
}
