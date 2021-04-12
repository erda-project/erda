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

package extmarketsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_cache_getExt(t *testing.T) {
	extCaches = &cache{Extensions: make(map[string]*apistructs.ExtensionVersion)}
	extCaches.Extensions["git"] = &apistructs.ExtensionVersion{Name: "git"}
	extCaches.Extensions["echo"] = &apistructs.ExtensionVersion{Name: "echo"}

	gitExt := extCaches.getExt("git")
	assert.NotNil(t, gitExt)
	assert.Equal(t, gitExt.Name, "git")

	javaExt := extCaches.getExt("java")
	assert.Nil(t, javaExt)
}

func Test_cache_updateExt(t *testing.T) {
	extCaches = &cache{Extensions: make(map[string]*apistructs.ExtensionVersion)}
	extCaches.Extensions["git"] = &apistructs.ExtensionVersion{Name: "git"}

	gitExt := extCaches.getExt("git")
	assert.NotNil(t, gitExt)

	extCaches.updateExt("git", apistructs.ExtensionVersion{Name: "git", Version: "1.0"})

	gitExt = extCaches.getExt("git")
	assert.NotNil(t, gitExt)
	assert.Equal(t, gitExt.Version, "1.0")
}
