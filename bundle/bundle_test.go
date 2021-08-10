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

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/http/httpclient"
)

func TestBundleOption(t *testing.T) {
	os.Setenv("CMDB_ADDR", "http://a.com")
	os.Setenv("DICEHUB_ADDR", "http://a.com")
	os.Setenv("EVENTBOX_ADDR", "http://a.com")
	os.Setenv("CMP_ADDR", "http://a.com")
	os.Setenv("ORCHESTRATOR_ADDR", "http://a.com")
	os.Setenv("SCHEDULER_ADDR", "http://a.com")
	os.Setenv("ADDON_PLATFORM_ADDR", "http://a.com")

	defer func() {
		os.Unsetenv("CMDB_ADDR")
		os.Unsetenv("DICEHUB_ADDR")
		os.Unsetenv("EVENTBOX_ADDR")
		os.Unsetenv("CMP_ADDR")
		os.Unsetenv("ORCHESTRATOR_ADDR")
		os.Unsetenv("SCHEDULER_ADDR")
		os.Unsetenv("ADDON_PLATFORM_ADDR")
	}()

	hc := httpclient.New()

	options := []Option{
		WithCMDB(),
		WithAddOnPlatform(),
		WithDiceHub(),
		WithEventBox(),
		WithCMP(),
		WithOrchestrator(),
		WithScheduler(),
		WithHTTPClient(hc),
	}

	b := New(options...)

	var (
		v   string
		err error
	)

	v, err = b.urls.CMDB()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.DiceHub()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.EventBox()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.CMP()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.Orchestrator()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.Scheduler()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

}
