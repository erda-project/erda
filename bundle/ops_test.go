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
