package kubernetes

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/bundle"
	clientset "github.com/erda-project/erda/pkg/clientgo/kubernetes"
	"github.com/erda-project/erda/pkg/httpserver"
)

const (
	NamespaceAll = "-"
)

type Endpoints struct {
	Bundle    *bundle.Bundle
	ClientSet map[string]kubernetes.Interface
}

type ListMetadata struct {
	Total int `json:"total"`
}

type ListSpec struct {
	List interface{} `json:"list"`
}

type Option func(endpoints *Endpoints)

func New(bdl *bundle.Bundle, ops ...Option) *Endpoints {
	ep := &Endpoints{
		Bundle:    bdl,
		ClientSet: make(map[string]kubernetes.Interface, 0),
	}
	for _, op := range ops {
		op(ep)
	}
	return ep
}

func (ep *Endpoints) Routers() (slice []httpserver.Endpoint) {
	eps := [][]httpserver.Endpoint{
		ep.ConfigMapRouters(),
		ep.DaemonSetRouters(),
		ep.NamespaceRouters(),
		ep.StatefulSetRouters(),
		ep.ServiceRouters(),
		ep.PodRouters(),
		ep.NodeRouters(),
	}

	for _, ep := range eps {
		slice = append(slice, ep...)
	}
	return
}

func (ep *Endpoints) GetClient(clusterName string) (client kubernetes.Interface, err error) {
	var (
		ok bool
	)
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}
	if client, ok = ep.ClientSet[clusterName]; !ok {
		cInfo, err := ep.Bundle.GetCluster(clusterName)
		if err != nil {
			logrus.Errorf("query cluster info failed, cluster:%s, err:%v", clusterName, err)
			return nil, err
		}
		if cInfo.SchedConfig == nil || cInfo.SchedConfig.MasterURL == "" {
			return nil, fmt.Errorf("empty inet address, cluster:%s", clusterName)
		}
		inetAddr := cInfo.SchedConfig.MasterURL

		client, err = clientset.NewKubernetesClientSet(inetAddr)
		if err != nil {
			logrus.Errorf("create k8s client failed, cluster:%s, err:%v", clusterName, err)
			return nil, err
		}
		ep.ClientSet[clusterName] = client
	}
	return ep.ClientSet[clusterName], nil
}

type PageInfo struct {
	PageSize int
	PageNo   int
}

func GetPageInfo(r *http.Request) (page PageInfo, err error) {
	page = PageInfo{}
	pageSizeStr := r.URL.Query().Get("pageSize")
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return
	}
	page.PageSize = pageSize
	page.PageNo = pageNo
	return
}

func (ep *Endpoints) GetFilterSlice(r *http.Request, list interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("the kind of the list is not a slice")
	}
	filterNameStr := r.URL.Query().Get("filterName")

	l := []interface{}{}
	for i := 0; i < v.Len(); i++ {
		name := v.Index(i).FieldByName("Name")
		if strings.Contains(name.String(), filterNameStr) {
			l = append(l, v.Index(i).Interface())
		}
	}
	return l, nil
}

func (ep *Endpoints) GetLimitSlice(r *http.Request, list interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("the kind of the list is not a slice")
	}
	pageInfo, err := GetPageInfo(r)
	if err != nil {
		return nil, err
	}

	var sv reflect.Value
	start := (pageInfo.PageNo - 1) * pageInfo.PageSize
	if start > v.Len() {
		return nil, fmt.Errorf("exceed limit of list length")
	}

	var end int
	if start+pageInfo.PageSize >= v.Len() {
		end = v.Len()
	} else {
		end = start + pageInfo.PageSize
	}
	sv = v.Slice(start, end)

	l := []interface{}{}
	for i := 0; i < sv.Len(); i++ {
		l = append(l, sv.Index(i).Interface())
	}

	return l, nil
}
