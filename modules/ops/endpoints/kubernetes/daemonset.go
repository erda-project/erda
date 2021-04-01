package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

type DaemonSetList struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) DaemonSetRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/daemonsets",
			Method:  http.MethodGet,
			Handler: ep.ListDaemonSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/daemonsets",
			Method:  http.MethodPost,
			Handler: ep.CreateDaemonSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/daemonsets/{daemonsetName}",
			Method:  http.MethodGet,
			Handler: ep.GetDaemonSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/daemonsets/{daemonsetName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateDaemonSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/daemonsets/{daemonsetName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteDaemonSet,
		},
	}
}

func (ep *Endpoints) ListDaemonSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clusterName string
	var ok bool
	var ns = metav1.NamespaceDefault
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if ns == NamespaceAll {
		ns = metav1.NamespaceAll
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	dsList, err := client.AppsV1().DaemonSets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get daemonset list error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, dsList.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	daemonsetList := DaemonSetList{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}
	return httpserver.OkResp(daemonsetList)
}

func (ep *Endpoints) GetDaemonSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clusterName string
	var ok bool
	var ns = metav1.NamespaceDefault
	var daemonsetName string
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if daemonsetName, ok = vars["daemonsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found daemonsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	ds, err := client.AppsV1().DaemonSets(ns).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get daemonset %s error: %s", daemonsetName, err.Error()))
	}
	return httpserver.OkResp(ds)
}

func (ep *Endpoints) CreateDaemonSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clusterName string
	var ok bool
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	ds := appsv1.DaemonSet{}
	err = json.NewDecoder(r.Body).Decode(&ds)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	if ds.Namespace == metav1.NamespaceNone {
		ds.Namespace = metav1.NamespaceDefault
	}

	newDS, err := client.AppsV1().DaemonSets(ds.Namespace).Create(ctx, &ds, metav1.CreateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("create daemonset error: %s", err.Error()))
	}

	return httpserver.OkResp(newDS)
}

func (ep *Endpoints) UpdateDaemonSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clusterName string
	var ok bool
	var ns = metav1.NamespaceDefault
	var daemonsetName string
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if daemonsetName, ok = vars["daemonsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found daemonsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	ds, err := client.AppsV1().DaemonSets(ns).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get daemonset %s error: %s", daemonsetName, err.Error()))
	}
	newDs := ds.DeepCopy()

	err = json.NewDecoder(r.Body).Decode(&newDs)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	if newDs.Name != daemonsetName {
		return errorresp.ErrResp(fmt.Errorf("orign daemonset name isn't equal new daemonset name"))
	}

	updateResp, err := client.AppsV1().DaemonSets(ns).Update(ctx, newDs, metav1.UpdateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("update daemonset %s error: %s", daemonsetName, err.Error()))
	}

	return httpserver.OkResp(updateResp)
}

func (ep *Endpoints) DeleteDaemonSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clusterName string
	var ok bool
	var ns = metav1.NamespaceDefault
	var daemonsetName string
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if daemonsetName, ok = vars["daemonsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found daemonsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	err = client.AppsV1().DaemonSets(ns).Delete(ctx, daemonsetName, metav1.DeleteOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("delete daemonset %s error: %s", daemonsetName, err.Error()))
	}
	return httpserver.OkResp(nil)
}
