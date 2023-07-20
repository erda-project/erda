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

package common_test

import (
	"net/http"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
)

func TestOn_On(t *testing.T) {
	var s = `
key: X-Ai-Proxy-Source
operator: =
value: erda.cloud
`
	var on common.On
	if err := yaml.Unmarshal([]byte(s), &on); err != nil {
		t.Fatal(err)
	}
	var h = http.Header{
		"X-Ai-Proxy-Source": {"erda.cloud"},
	}
	ok, err := on.On(h)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ok)
	if !ok {
		t.Fatal("it should be ok")
	}
}
