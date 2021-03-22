package engines

import (
	"github.com/erda-project/erda/pkg/clientgo"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/executors"
)

type LocalEngine struct {
	istioctl.DefaultEngine
}

func NewLocalEngine(addr string) (*LocalEngine, error) {
	client, err := clientgo.New(addr)
	if err != nil {
		return nil, err
	}
	authN := &executors.AuthNExecutor{}
	authN.SetIstioClient(client.CustomClient)
	return &LocalEngine{
		DefaultEngine: istioctl.NewDefaultEngine(authN),
	}, nil
}
