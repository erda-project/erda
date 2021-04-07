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

package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/aliyun/alibaba-cloud-sdk-go/services/cloudapi"
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/require"
//)
//
//var (
//	clusterName = "terminus-dev"
//)
//
//func TestBundle_DoRemoteAliyunAction(t *testing.T) {
//	os.Setenv("OPS_ADDR", "ops.default.svc.cluster.local:9027")
//	defer func() {
//		os.Unsetenv("OPS_ADDR")
//	}()
//	logrus.SetOutput(os.Stdout)
//	b := New(WithOps())
//	resp := cloudapi.CreateDescribeZonesResponse()
//	err := b.DoRemoteAliyunAction("1", "terminus-dev", cloudapi.GetEndpointType(),
//		cloudapi.GetEndpointMap(), cloudapi.CreateDescribeZonesRequest(), resp)
//	logrus.Error(resp)
//	require.NoError(t, err)
//}
