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
	"testing"

	"bou.ke/monkey"

	"github.com/alecthomas/assert"
)

func TestGetDomain(t *testing.T) {
	host := "www.sfsfsf.erda.cloud"
	confDomain := ".terminus.io,.erda.cloud"
	domain, err := GetDomain(host, confDomain)
	assert.NoError(t, err)
	assert.Equal(t, ".erda.cloud", domain)
}

func TestGetUCRedirectHost(t *testing.T) {
	referer := "https://erda.cloud"
	guard := monkey.Patch(UCRedirectHost, func() string {
		return "openapi.dev.terminus.io,openapi.erda.cloud"
	})
	defer guard.Unpatch()
	host := GetUCRedirectHost(referer)
	assert.Equal(t, "", host)
}
