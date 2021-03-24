package pipelinesvc

import (
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

func TestValidateCreateRequest(t *testing.T) {
	os.Setenv("SCHEDULER_ADDR", "scheduler.default.svc.cluster.local:9091")
	svc := PipelineSvc{
		bdl: bundle.New(
			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))),
			bundle.WithScheduler(),
		),
	}

	req := apistructs.PipelineCreateRequestV2{
		PipelineYml:    "1.yml",
		ClusterName:    "local",
		PipelineSource: apistructs.PipelineSourceQA,
		IdentityInfo:   apistructs.IdentityInfo{InternalClient: "local"},
		Labels: map[string]string{
			"1": "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789": "1",
			"key": "v",
		},
		NormalLabels: map[string]string{
			"1": "value",
		},
	}
	err := svc.validateCreateRequest(&req)
	assert.NoError(t, err)
	spew.Dump(req)
}
