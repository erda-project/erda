package bundle

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestBundle_GetNotifyConfigMS(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	defer func() {
		os.Unsetenv("CMDB_ADDR")
	}()
	bdl := New(WithCMDB())
	cfg, err := bdl.GetNotifyConfigMS("2", "1")
	assert.NoError(t, err)
	assert.Equal(t, false, cfg)
}

func TestBundle_NotifyList(t *testing.T) {
	os.Setenv("MONITOR_ADDR", "monitor.default.svc.cluster.local:7096")
	defer func() {
		os.Unsetenv("MONITOR_ADDR")
	}()
	bdl := New(WithMonitor())
	req := apistructs.NotifyPageRequest{
		ScopeId: "18",
		Scope:   "app",
		UserId:  "2",
		OrgId:   "1",
	}
	cfg, err := bdl.NotifyList(req)
	assert.NoError(t, err)
	fmt.Printf("%+v", cfg)
}
