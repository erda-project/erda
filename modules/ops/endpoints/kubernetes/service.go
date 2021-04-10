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

type Service struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) ServiceRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/services",
			Method:  http.MethodGet,
			Handler: ep.ListService,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/services/{serviceName}",
			Method:  http.MethodGet,
			Handler: ep.GetService,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/services",
			Method:  http.MethodPost,
			Handler: ep.CreateService,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/services/{serviceName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateService,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/services/{serviceName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteService,
		},
	}
}

func (ep *Endpoints) ListService(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
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

	list, err := client.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get service list failed, error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, list.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	data := Service{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}
	return httpserver.OkResp(data)
}

func (ep *Endpoints) GetService(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		ok          bool
		clusterName string
		serviceName string
		client      kubernetes.Interface
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if serviceName, ok = vars["serviceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configmap name"))
	}

	client, err = ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	service, err := client.CoreV1().Services(ns).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get service failed, error: %s", err.Error()))
	}

	data := service
	return httpserver.OkResp(data)
}

func (ep *Endpoints) CreateService(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	service := corev1.Service{}
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}
	if service.Namespace == metav1.NamespaceNone {
		service.Namespace = metav1.NamespaceDefault
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().Services(service.Namespace).Create(ctx, &service, metav1.CreateOptions{})
	if err != nil {
		err := fmt.Errorf("create service failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) UpdateService(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		serviceName string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if serviceName, ok = vars["serviceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	service := corev1.Service{}
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}

	if serviceName != service.Name {
		return errorresp.ErrResp(fmt.Errorf("service name connot be modified, origin:%s, modified:%s", serviceName, service.Name))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	_, err = client.CoreV1().Services(ns).Update(ctx, &service, metav1.UpdateOptions{})
	if err != nil {
		err := fmt.Errorf("update service failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) DeleteService(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		ok          bool
		clusterName string
		serviceName string
		ns          = metav1.NamespaceDefault
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if serviceName, ok = vars["serviceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found configMapName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	err = client.CoreV1().Services(ns).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		err := fmt.Errorf("delete service failed, cluster:%s, err:%v", clusterName, err)
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}
