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
