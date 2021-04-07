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
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

type NamespaceList struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) NamespaceRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces",
			Method:  http.MethodGet,
			Handler: ep.ListNamespace,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces",
			Method:  http.MethodPost,
			Handler: ep.CreateNamespace,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}",
			Method:  http.MethodGet,
			Handler: ep.GetNamespace,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateNamespace,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteNamespace,
		},
	}
}

func (ep *Endpoints) ListNamespace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName string
		ok          bool
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get namespace list error: %s", err.Error()))
	}

	// remove default ns
	var (
		index     int
		defaultNS = corev1.Namespace{}
	)
	for index = range nsList.Items {
		if nsList.Items[index].Name == metav1.NamespaceDefault {
			defaultNS = nsList.Items[index]
			break
		}
	}
	nsList.Items = append(nsList.Items[:index], nsList.Items[index+1:]...)

	sort.SliceIsSorted(nsList.Items, func(i, j int) bool {
		if nsList.Items[i].Name < nsList.Items[j].Name {
			return true
		}
		return false
	})

	nsList.Items = append([]corev1.Namespace{defaultNS}, nsList.Items...)

	filterList, err := ep.GetFilterSlice(r, nsList.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	namespaceList := NamespaceList{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: filterList},
	}
	return httpserver.OkResp(namespaceList)
}

func (ep *Endpoints) GetNamespace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName   string
		ok            bool
		namespaceName string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if namespaceName, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get namespace %s error: %s", namespaceName, err.Error()))
	}
	return httpserver.OkResp(ns)
}

func (ep *Endpoints) CreateNamespace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName string
		ok          bool
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	ns := corev1.Namespace{}
	err = json.NewDecoder(r.Body).Decode(&ns)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json body to namespaces error: %s", err.Error()))
	}

	nsResp, err := client.CoreV1().Namespaces().Create(ctx, &ns, metav1.CreateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("create namespace error: %s", err.Error()))
	}
	return httpserver.OkResp(nsResp)
}

func (ep *Endpoints) UpdateNamespace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName   string
		ok            bool
		namespaceName string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if namespaceName, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get namespaces %s error: %s", namespaceName, err.Error()))
	}
	newNamespace := ns.DeepCopy()

	err = json.NewDecoder(r.Body).Decode(&newNamespace)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json body to namespaces error: %s", err.Error()))
	}

	nsResp, err := client.CoreV1().Namespaces().Update(ctx, newNamespace, metav1.UpdateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("update namespace %s error: %s", namespaceName, err.Error()))
	}
	return httpserver.OkResp(nsResp)
}

func (ep *Endpoints) DeleteNamespace(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName   string
		ok            bool
		namespaceName string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if namespaceName, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	err = client.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("delete namespace %s error: %s", namespaceName, err.Error()))
	}
	return httpserver.OkResp(nil)
}
