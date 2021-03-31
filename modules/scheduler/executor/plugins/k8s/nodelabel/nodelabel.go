package nodelabel

import (
	"bytes"
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/httpclient"
)

type NodeLabel struct {
	addr   string
	client *httpclient.HTTPClient
}

func New(addr string, client *httpclient.HTTPClient) *NodeLabel {
	return &NodeLabel{
		addr:   addr,
		client: client,
	}
}

func (n *NodeLabel) List() (*v1.NodeList, error) {
	var body v1.NodeList
	path := "/api/v1/nodes"
	r, err := n.client.Get(n.addr).
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

func (n *NodeLabel) Get(host string) (map[string]string, error) {
	var body v1.Node
	path := fmt.Sprintf("/api/v1/nodes/%s", host)
	r, err := n.client.Get(n.addr).
		Path(path).
		Do().
		JSON(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to get node labels, host: %v, err: %v", host, err)
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("failed to get node labels, host: %v, statuscode:%d", host, r.StatusCode())
	}
	return body.
		ObjectMeta.Labels, nil
}

func (n *NodeLabel) Set(labels map[string]*string, host string) error {
	var req struct {
		Metadata struct {
			Labels map[string]*string `json:"labels"` // Use '*string' to cover 'null' case
		} `json:"metadata"`
	}
	var body bytes.Buffer

	req.Metadata.Labels = map[string]*string{}
	for k, v := range labels {
		req.Metadata.Labels[k] = v
	}
	path := fmt.Sprintf("/api/v1/nodes/%s", host)
	r, err := n.client.Patch(n.addr).
		Path(path).
		JSONBody(req).
		Header("Content-Type", "application/merge-patch+json").
		Do().
		Body(&body)
	if err != nil {
		return fmt.Errorf("failed to update node labels, host: %v, err: %v", host, err)
	}
	if !r.IsOK() {
		return fmt.Errorf("failed to update node labels, host:%v, statuscode: %v, body: %v", host, r.StatusCode(), body.String())
	}
	return nil
}
