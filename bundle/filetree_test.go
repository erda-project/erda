package bundle

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"
)

func TestBundle_ListGittarFileTreeNodes(t *testing.T) {
	hc := httpclient.New(httpclient.WithEnableAutoRetry(false))
	os.Setenv("GITTAR_ADAPTOR_ADDR", "gittar-adaptor.default.svc.cluster.local:1086")
	bdl := New(WithGittarAdaptor(), WithHTTPClient(hc))
	begin := time.Now()
	nodes, err := bdl.ListGittarFileTreeNodes(apistructs.UnifiedFileTreeNodeListRequest{
		Scope:        "project-app",
		ScopeID:      "59",
		Pinode:       "NTkvMjI=",
		IdentityInfo: apistructs.IdentityInfo{UserID: "1"},
	}, 1)
	end := time.Now()
	fmt.Println(end.Sub(begin))
	assert.NoError(t, err)
	fmt.Println(len(nodes))
}
