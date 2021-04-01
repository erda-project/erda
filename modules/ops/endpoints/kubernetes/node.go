package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/k8s/node/drain"
)

func (ep *Endpoints) NodeRouters() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{
			Path:    "/apis/clusters/{clusterName}/nodes/taint",
			Method:  http.MethodPut,
			Handler: ep.UpdateNodeTaints,
		},
		{
			Path:    "/apis/clusters/{clusterName}/nodes/drain",
			Method:  http.MethodPost,
			Handler: ep.DrainNode,
		},
	}
}

type NodeMeta struct {
	Name string `json:"name"`
}

type NodeSpec struct {
	IP string `json:"IP"`
}

type Node struct {
	NodeMeta
	NodeSpec
}

type UpdateNodeTagsRequest struct {
	Taints []corev1.Taint `json:"taints"`
	Nodes  []Node         `json:"nodes"`
}

func (ep *Endpoints) UpdateNodeTaints(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName string
		ok          bool
	)

	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	req := UpdateNodeTagsRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	for _, v := range req.Nodes {
		if v.Name == "" {
			return errorresp.ErrResp(fmt.Errorf("empty node name"))
		}
		// patch
		taints := corev1.Node{
			Spec: corev1.NodeSpec{Taints: req.Taints},
		}
		data, _ := json.Marshal(taints)
		_, err = client.CoreV1().Nodes().Patch(ctx, v.Name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
		if err != nil {
			errNew := fmt.Errorf("update node [%s] failed, err:%v", v.Name, err)
			logrus.Errorf(errNew.Error())
			return errorresp.ErrResp(errNew)
		}
	}
	return httpserver.OkResp(nil)
}

func (ep *Endpoints) DrainNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusterName string
		ok          bool
	)
	if clusterName, ok = vars["clusterName"]; !ok {
		return errorresp.ErrResp(fmt.Errorf("not found clusterName"))
	}

	req := apistructs.DrainNodeRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("request decode failed, err:%v", err))
	}

	client, err := ep.GetClient(clusterName)
	if err != nil {
		return errorresp.ErrResp(fmt.Errorf("get client failed, err:%v", err))
	}

	err = drain.DrainNode(ctx, client, req)
	if err != nil {
		logrus.Errorf("drain node failed, request: %v, error:%v", req, err)
		return errorresp.ErrResp(fmt.Errorf("drain node failed, err:%v", err))
	}
	return httpserver.OkResp(nil)
}
