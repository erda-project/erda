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

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

type Pod struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) PodRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/pods",
			Method:  http.MethodGet,
			Handler: ep.ListPod,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/pods/{podName}",
			Method:  http.MethodGet,
			Handler: ep.GetPod,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/pods",
			Method:  http.MethodPost,
			Handler: ep.CreatePod,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/pods/{podName}",
			Method:  http.MethodPut,
			Handler: ep.UpdatePod,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/pods/{podName}",
			Method:  http.MethodDelete,
			Handler: ep.DeletePod,
		},
	}
}

func (ep *Endpoints) ListPod(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		ok            bool
		clusterName   string
		client        kubernetes.Interface
		ns            = metav1.NamespaceDefault
		labelSelector string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if ns == NamespaceAll {
		ns = metav1.NamespaceAll
	}

	if v, ok := r.URL.Query()["labelSelector"]; ok {
		labelSelector, _ = url.QueryUnescape(v[0])
	}

	client, err = ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	podlist, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get pod list failed, error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, podlist.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	data := Pod{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}
	return httpserver.OkResp(data)
}

func (ep *Endpoints) GetPod(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		ok          bool
		clusterName string
		podName     string
		client      kubernetes.Interface
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if podName, ok = vars["podName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configmap name"))
	}

	client, err = ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	pod, err := client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get pod failed, error: %s", err.Error()))
	}

	data := pod
	return httpserver.OkResp(data)
}

func (ep *Endpoints) CreatePod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	pod := corev1.Pod{}
	err := json.NewDecoder(r.Body).Decode(&pod)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}
	if pod.Namespace == metav1.NamespaceNone {
		pod.Namespace = metav1.NamespaceDefault
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().Pods(ns).Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		err := fmt.Errorf("create pod failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) UpdatePod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		podName     string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if podName, ok = vars["podName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	pod := corev1.Pod{}
	err := json.NewDecoder(r.Body).Decode(&pod)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}

	if podName != pod.Name {
		return errorresp.ErrResp(fmt.Errorf("pod name connot be modified, origin:%s, modified:%s", podName, pod.Name))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().Pods(ns).Update(ctx, &pod, metav1.UpdateOptions{})
	if err != nil {
		err := fmt.Errorf("update pod failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) DeletePod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		podName     string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if podName, ok = vars["podName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	err = client.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		err := fmt.Errorf("delete pod failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}
