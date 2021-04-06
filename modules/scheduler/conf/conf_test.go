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
