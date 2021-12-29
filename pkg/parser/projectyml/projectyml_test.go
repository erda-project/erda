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

package projectyml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	data := `version: "1.1"
applications:
    - name: app1
      display_name: app1
      logo: http://xxx/erda.png
      desc: aaa
      repo_config: 
        type: external`
	yml, err := New([]byte(data))
	assert.NoError(t, err)
	fmt.Println(yml)
	assert.Equal(t, "app1", yml.s.Applications[0].Name)
	assert.Equal(t, "http://xxx/erda.png", yml.s.Applications[0].Logo)
	assert.Equal(t, "external", yml.s.Applications[0].RepoConfig.Type)
}
