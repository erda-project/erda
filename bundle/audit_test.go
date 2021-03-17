package bundle

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
)

func TestBundle_CreateEvent(t *testing.T) {
	os.Setenv("CMDB_ADDR", "cmdb.marathon.l4lb.thisdcos.directory:9093")
	defer func() {
		os.Unsetenv("CMDB_ADDR")
	}()
	b := New(WithCMDB())
	tm := strconv.FormatInt(time.Now().Unix(), 10)
	audit := apistructs.Audit{
		UserID:       "1000008",
		ScopeType:    "app",
		ScopeID:      2,
		AppID:        4,
		ProjectID:    2,
		OrgID:        2,
		Context:      map[string]interface{}{"appId": "4", "projectId": "2"},
		TemplateName: "createApp",
		AuditLevel:   "p3",
		Result:       "success",
		ErrorMsg:     "",
		StartTime:    tm,
		EndTime:      tm,
		ClientIP:     "1.1.1.1",
		UserAgent:    "chrom",
	}
	err := b.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: audit})
	require.NoError(t, err)
}
