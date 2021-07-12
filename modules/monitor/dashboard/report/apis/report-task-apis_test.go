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

package apis

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/discover"
)

func Test_provider_createFQDN(t *testing.T) {
	p := &provider{
		Cfg: &config{
			DiceNameSpace: "project-xxx",
		},
	}
	ass := assert.New(t)

	os.Setenv("MONITOR_ADDR", "monitor:7096")
	addr, err := p.createFQDN(discover.Monitor())
	ass.Nil(err)
	assert.Equal(t, "monitor.project-xxx:7096", addr)

	os.Setenv("MONITOR_ADDR", "monitor")
	addr, err = p.createFQDN(discover.Monitor())
	ass.Error(err)

	os.Setenv("MONITOR_ADDR", "http://monitor:7096")
	addr, err = p.createFQDN(discover.Monitor())
	ass.Error(err)
}
