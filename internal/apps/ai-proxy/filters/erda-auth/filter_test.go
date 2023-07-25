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

package erda_auth_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	erda_auth "github.com/erda-project/erda/internal/apps/ai-proxy/filters/erda-auth"
)

func TestConfig(t *testing.T) {
	var y = `
on:
  - key: X-Ai-Proxy-Source
    operator: =
    value: erda.cloud
credential:
  name: erda.cloud
  platform: erda
  provider: ""
  providerInstanceId: ""
`
	var cfg erda_auth.Config
	if err := yaml.Unmarshal([]byte(y), &cfg); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cfg)
}
