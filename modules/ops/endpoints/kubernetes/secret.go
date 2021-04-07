package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

type SecretList struct {
	ListMetadata
	ListSpec
}

func (ep *Endpoints) SecretRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/secrets",
			Method:  http.MethodGet,
			Handler: ep.ListSecret,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/secrets",
			Method:  http.MethodPost,
			Handler: ep.CreateSecret,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/secrets/{secretName}",
			Method:  http.MethodGet,
			Handler: ep.GetSecret,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/secrets/{secretName}",
			Method:  http.MethodPut,
			Handler: ep.UpdateSecret,
		},
		{
			Path:    "/apis/clusters/{clusterName}/namespaces/{namespaceName}/secrets/{secretName}",
			Method:  http.MethodDelete,
			Handler: ep.DeleteSecret,
		},
	}
}

func (ep *Endpoints) ListSecret(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
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

	scList, err := client.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get secret list error: %s", err.Error()))
	}

	filterList, err := ep.GetFilterSlice(r, scList.Items)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("filter list error: %s", err.Error()))
	}

	limitList, err := ep.GetLimitSlice(r, filterList)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("limit list error: %s", err.Error()))
	}

	secretList := SecretList{
		ListMetadata{Total: len(filterList)},
		ListSpec{List: limitList},
	}

	return httpserver.OkResp(secretList)
}

func (ep *Endpoints) GetSecret(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		clusterName string
		ok          bool
		ns          = metav1.NamespaceDefault
		secretName  string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if secretName, ok = vars["secretName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found secretName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	sc, err := client.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get secret %s error: %s", secretName, err.Error()))
	}

	return httpserver.OkResp(sc)
}

func (ep *Endpoints) CreateSecret(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
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

	sc := corev1.Secret{}

	err = json.NewDecoder(r.Body).Decode(&sc)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	if sc.Namespace == metav1.NamespaceNone {
		sc.Namespace = metav1.NamespaceDefault
	}

	scResp, err := client.CoreV1().Secrets(sc.Namespace).Create(ctx, &sc, metav1.CreateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("create secret error: %s", err.Error()))
	}

	return httpserver.OkResp(scResp)
}

func (ep *Endpoints) UpdateSecret(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		clusterName string
		ok          bool
		ns          = metav1.NamespaceDefault
		secretName  string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if secretName, ok = vars["secretName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found secretName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	sc, err := client.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get secret %s error: %s", secretName, err.Error()))
	}

	newSc := sc.DeepCopy()

	err = json.NewDecoder(r.Body).Decode(&newSc)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("unmarshal json from request error: %s", err.Error()))
	}

	updateResp, err := client.CoreV1().Secrets(ns).Update(ctx, newSc, metav1.UpdateOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("update daemonset %s error: %s", secretName, err.Error()))
	}

	return httpserver.OkResp(updateResp)
}

func (ep *Endpoints) DeleteSecret(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	var (
		clusterName string
		ok          bool
		ns          = metav1.NamespaceDefault
		secretName  string
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	if ns, ok = vars["namespaceName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found namespaceName"))
	}

	if secretName, ok = vars["secretName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found secretName"))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	err = client.CoreV1().Secrets(ns).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("delete secret %s error: %s", secretName, err.Error()))
	}

	return httpserver.OkResp(nil)
}
