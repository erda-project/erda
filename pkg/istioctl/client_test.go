package istioctl

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/clientgo"
	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func TestNewIstioClient(t *testing.T) {
	restclient.SetInetAddr("netportal.default.svc.cluster.local")
	client, _ := clientgo.New("inet://dev.terminus.io/kubernetes.default.svc.cluster.local")
	// if err != nil {
	// 	panic(err)
	// }
	ns := "foo"
	var dr interface{}
	err := client.CustomClient.NetworkingV1alpha3().DestinationRules(ns).Delete(context.TODO(), "httpbin", metav1.DeleteOptions{})
	if err != nil {
		assert.Equal(t, errors.IsNotFound(err), true)
	}
	fmt.Printf("%+v\n, err:%+v", dr, err)
}
