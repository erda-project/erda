// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipelinesvc

import (
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpclient"
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
