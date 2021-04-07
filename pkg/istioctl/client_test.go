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

package istioctl

//import (
//	"context"
//	"fmt"
//	"testing"
//
//	"gotest.tools/assert"
//	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
//	"github.com/erda-project/erda/pkg/clientgo"
//	"github.com/erda-project/erda/pkg/clientgo/restclient"
//)
//
//func TestNewIstioClient(t *testing.T) {
//	restclient.SetInetAddr("netportal.default.svc.cluster.local")
//	client, _ := clientgo.New("inet://dev.terminus.io/kubernetes.default.svc.cluster.local")
//	// if err != nil {
//	// 	panic(err)
//	// }
//	ns := "foo"
//	var dr interface{}
//	err := client.CustomClient.NetworkingV1alpha3().DestinationRules(ns).Delete(context.TODO(), "httpbin", metav1.DeleteOptions{})
//	if err != nil {
//		assert.Equal(t, errors.IsNotFound(err), true)
//	}
//	fmt.Printf("%+v\n, err:%+v", dr, err)
//}
