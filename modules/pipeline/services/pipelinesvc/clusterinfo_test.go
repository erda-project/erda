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

//import (
//	"fmt"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/bundle"
//	"github.com/erda-project/erda/pkg/httpclient"
//)
//
//func TestPipelineSvc_retryQueryClusterInfo(t *testing.T) {
//	os.Setenv("SCHEDULER_ADDR", "scheduler.default.svc.cluster.local:9091")
//	svc := PipelineSvc{
//		bdl: bundle.New(
//			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))),
//			bundle.WithScheduler(),
//		),
//	}
//
//	clusterInfo, err := svc.retryQueryClusterInfo("terminus-test", 1)
//	assert.NoError(t, err)
//	fmt.Println(clusterInfo)
//}
