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
