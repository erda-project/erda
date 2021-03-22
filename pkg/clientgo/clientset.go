package clientgo

import (
	"github.com/erda-project/erda/pkg/clientgo/customclient"
	"github.com/erda-project/erda/pkg/clientgo/kubernetes"
)

type ClientSet struct {
	K8sClient    *kubernetes.Clientset
	CustomClient *customclient.Clientset
}

func New(addr string) (*ClientSet, error) {
	var cs ClientSet
	var err error
	cs.K8sClient, err = kubernetes.NewKubernetesClientSet(addr)
	if err != nil {
		return nil, err
	}
	cs.CustomClient, err = customclient.NewCustomClientSet(addr)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
