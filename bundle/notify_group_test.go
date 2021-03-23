package bundle

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundle_GetNotifyConfig(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	defer func() {
		os.Unsetenv("CMDB_ADDR")
	}()
	bdl := New(WithCMDB())

	cfg, err := bdl.GetNotifyConfig("1", "2")
	assert.NoError(t, err)
	fmt.Printf("%+v\n", *cfg.Config)
}
