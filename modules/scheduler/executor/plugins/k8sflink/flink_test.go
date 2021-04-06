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

package k8sflink

//
//import (
//	"context"
//	"fmt"
//	"testing"
//
//	corev1 "k8s.io/api/core/v1"
//
//	"gotest.tools/assert"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
//	"github.com/erda-project/erda/pkg/clientgo"
//	"github.com/erda-project/erda/pkg/clientgo/restclient"
//)
//
//func TestNewFlinkClient(t *testing.T) {
//	restclient.SetInetAddr("netportal.default.svc.cluster.local")
//	client, _ := clientgo.New("inet://dev.terminus.io/kubernetes.default.svc.cluster.local")
//
//	newNS := corev1.Namespace{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "dd-test",
//		},
//	}
//	n, err := client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), &newNS, metav1.CreateOptions{})
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	ns := "default"
//
//	defaultSC, err := client.K8sClient.CoreV1().Secrets(ns).Get(context.TODO(), AliyunPullSecret, metav1.GetOptions{})
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	sc := corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      AliyunPullSecret,
//			Namespace: n.Name,
//		},
//		Data: defaultSC.Data,
//		Type: defaultSC.Type,
//	}
//	_, e := client.K8sClient.CoreV1().Secrets(n.Name).Create(context.TODO(), &sc, metav1.CreateOptions{})
//	if e != nil {
//		fmt.Println(e)
//		return
//	}
//
//	list, err := client.CustomClient.
//		FlinkoperatorV1beta1().
//		FlinkClusters(ns).
//		List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		fmt.Println(err)
//	}
//	assert.Equal(t, len(list.Items), 0)
//	err = client.CustomClient.
//		FlinkoperatorV1beta1().
//		FlinkClusters(ns).
//		Delete(context.TODO(), "flinksessioncluster", metav1.DeleteOptions{})
//	if err != nil {
//		fmt.Println(err)
//	}
//	list, err = client.CustomClient.
//		FlinkoperatorV1beta1().
//		FlinkClusters(ns).
//		List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		fmt.Println(err)
//	}
//	assert.Equal(t, len(list.Items), 0)
//	fmt.Printf("%+v\n", list.Items)
//}
