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

type StatefulSetList struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) StatefulSetRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/statefulsets",
			Method:  http.MethodGet,
			Handler: ep.ListStatefulSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/statefulsets",
			Method:  http.MethodPost,
			Handler: ep.CreateStatefulSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/statefulsets/{statefulsetName}",
			Method:  http.MethodGet,
			Handler: ep.GetStatefulSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/statefulsets/{statefulsetName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateStatefulSet,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/statefulsets/{statefulsetName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteStatefulSet,
		},
	}
}

func (ep *Endpoints) ListStatefulSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName string
		ok          bool
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

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	stsList, err := client.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get statefulset list error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, stsList.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	statefulsetList := StatefulSetList{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}
	return httpserver.OkResp(statefulsetList)
}

func (ep *Endpoints) GetStatefulSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName    string
		ok             bool
		ns             = metav1.NamespaceDefault
		statfulsetName string
	)
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if statfulsetName, ok = vars["statefulsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found statfulsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	sts, err := client.AppsV1().StatefulSets(ns).Get(ctx, statfulsetName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get statfulset %s error: %s", statfulsetName, err.Error()))
	}
	return httpserver.OkResp(sts)
}

func (ep *Endpoints) CreateStatefulSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
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

	sts := appsv1.StatefulSet{}
	err = json.NewDecoder(r.Body).Decode(&sts)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	if sts.Namespace == metav1.NamespaceNone {
		sts.Namespace = metav1.NamespaceDefault
	}

	newDS, err := client.AppsV1().StatefulSets(sts.Namespace).Create(ctx, &sts, metav1.CreateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("create statfulset error: %s", err.Error()))
	}

	return httpserver.OkResp(newDS)
}

func (ep *Endpoints) UpdateStatefulSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName    string
		ok             bool
		ns             = metav1.NamespaceDefault
		statfulsetName string
	)
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if statfulsetName, ok = vars["statefulsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found statfulsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	sts, err := client.AppsV1().StatefulSets(ns).Get(ctx, statfulsetName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get statfulset %s error: %s", statfulsetName, err.Error()))
	}
	newSts := sts.DeepCopy()

	err = json.NewDecoder(r.Body).Decode(&newSts)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	if newSts.Name != statfulsetName {
		return errorresp.ErrResp(fmt.Errorf("orign statfulset name isn't equal new statfulset name"))
	}

	updateResp, err := client.AppsV1().StatefulSets(ns).Update(ctx, newSts, metav1.UpdateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("update daemonset %s error: %s", statfulsetName, err.Error()))
	}

	return httpserver.OkResp(updateResp)
}

func (ep *Endpoints) DeleteStatefulSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName    string
		ok             bool
		ns             = metav1.NamespaceDefault
		statfulsetName string
	)
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}
	if statfulsetName, ok = vars["statefulsetName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found statfulsetName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	err = client.AppsV1().StatefulSets(ns).Delete(ctx, statfulsetName, metav1.DeleteOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("delete statfulsetName %s error: %s", statfulsetName, err.Error()))
	}
	return httpserver.OkResp(nil)
}
