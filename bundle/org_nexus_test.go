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

package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestBundle_GetNexusOrgDockerCredentialByImage(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "localhost:9093")
//	defer os.Unsetenv("CMDB_ADDR")
//	bdl := New(WithCMDB())
//	user, err := bdl.GetNexusOrgDockerCredentialByImage(1, "nginx")
//	assert.NoError(t, err)
//	assert.Nil(t, user)
//
//	user, err = bdl.GetNexusOrgDockerCredentialByImage(1, "docker-hosted-org-1-nexus-sys.dev.terminus.io/terminus-dice-dev/test-release-cross-cluster:dockerfile-1593662620685244179")
//	assert.NoError(t, err)
//	assert.NotNil(t, user)
//	spew.Dump(user)
//}
