package pipelinesvc

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

func TestPipelineSvc_retryQueryClusterInfo(t *testing.T) {
	os.Setenv("SCHEDULER_ADDR", "scheduler.default.svc.cluster.local:9091")
	svc := PipelineSvc{
		bdl: bundle.New(
			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))),
			bundle.WithScheduler(),
		),
	}

	clusterInfo, err := svc.retryQueryClusterInfo("terminus-test", 1)
	assert.NoError(t, err)
	fmt.Println(clusterInfo)
}
