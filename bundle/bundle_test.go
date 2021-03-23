package bundle

import (
	"os"
	"testing"

	"github.com/erda-project/erda/pkg/httpclient"

	"github.com/stretchr/testify/assert"
)

func TestBundleOption(t *testing.T) {
	os.Setenv("CMDB_ADDR", "http://a.com")
	os.Setenv("DICEHUB_ADDR", "http://a.com")
	os.Setenv("EVENTBOX_ADDR", "http://a.com")
	os.Setenv("OPS_ADDR", "http://a.com")
	os.Setenv("ORCHESTRATOR_ADDR", "http://a.com")
	os.Setenv("SCHEDULER_ADDR", "http://a.com")
	os.Setenv("ADDON_PLATFORM_ADDR", "http://a.com")

	defer func() {
		os.Unsetenv("CMDB_ADDR")
		os.Unsetenv("DICEHUB_ADDR")
		os.Unsetenv("EVENTBOX_ADDR")
		os.Unsetenv("OPS_ADDR")
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
		WithOps(),
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

	v, err = b.urls.Ops()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.Orchestrator()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

	v, err = b.urls.Scheduler()
	assert.Equal(t, v, "http://a.com")
	assert.Nil(t, err)

}
