// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package proxy

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestReplaceServiceName(t *testing.T) {
	os.Setenv("ERDA_SYSTEM_FQDN", "erda-system.svc.cluster.local")
	defer os.Unsetenv("ERDA_SYSTEM_FQDN")
	result := replaceServiceName(os.Getenv("ERDA_SYSTEM_FQDN"), "openapi.default.svc.cluster.local")
	assert.Equal(t, "openapi.erda-system.svc.cluster.local", result)
}
