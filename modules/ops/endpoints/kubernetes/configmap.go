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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

type ConfigMap struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) ConfigMapRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/configmaps",
			Method:  http.MethodGet,
			Handler: ep.ListConfigMap,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/configmaps/{cmName}",
			Method:  http.MethodGet,
			Handler: ep.GetConfigMap,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/configmaps",
			Method:  http.MethodPost,
			Handler: ep.CreateConfigMap,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/configmaps/{cmName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateConfigMap,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/configmaps/{cmName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteConfigMap,
		},
	}
}

func (ep *Endpoints) ListConfigMap(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		ok          bool
		clusterName string
		client      kubernetes.Interface
		ns          = metav1.NamespaceDefault
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

	client, err = ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	cmList, err := client.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get cm list failed, error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, cmList.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	data := ConfigMap{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}
	return httpserver.OkResp(data)
}

func (ep *Endpoints) GetConfigMap(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		ok          bool
		clusterName string
		cmName      string
		client      kubernetes.Interface
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if cmName, ok = vars["cmName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configmap name"))
	}

	client, err = ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	cm, err := client.CoreV1().ConfigMaps(ns).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get cm failed, error: %s", err.Error()))
	}

	data := cm
	return httpserver.OkResp(data)
}

func (ep *Endpoints) CreateConfigMap(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	cm := corev1.ConfigMap{}
	err := json.NewDecoder(r.Body).Decode(&cm)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}
	if cm.Namespace == metav1.NamespaceNone {
		cm.Namespace = metav1.NamespaceDefault
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, &cm, metav1.CreateOptions{})
	if err != nil {
		err := fmt.Errorf("create cm failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) UpdateConfigMap(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		cmName      string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if cmName, ok = vars["cmName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	cm := corev1.ConfigMap{}
	err := json.NewDecoder(r.Body).Decode(&cm)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}

	if cmName != cm.Name {
		return errorresp.ErrResp(fmt.Errorf("config map name connot be modified, origin:%s, modified:%s", cmName, cm.Name))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().ConfigMaps(ns).Update(ctx, &cm, metav1.UpdateOptions{})
	if err != nil {
		err := fmt.Errorf("update cm failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) DeleteConfigMap(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		cmName      string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if cmName, ok = vars["cmName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	err = client.CoreV1().ConfigMaps(ns).Delete(ctx, cmName, metav1.DeleteOptions{})
	if err != nil {
		err := fmt.Errorf("delete cm failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}
