package node

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/httpclient"
)

type Node struct {
	addr   string
	client *httpclient.HTTPClient
}

func New(addr string, client *httpclient.HTTPClient) *Node {
	return &Node{
		addr:   addr,
		client: client,
	}
}

func (n *Node) List() (*v1.NodeList, error) {
	var body v1.NodeList
	path := "/api/v1/nodes"
	r, err := n.client.Get(n.addr, httpclient.RetryOption{}).
		Path(path).
		Do().
		JSON(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodelist, err: %v", err)
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get nodelist, statuscode: %d, body: %s", r.StatusCode(), r.Body())
	}
	return &body, nil
}
