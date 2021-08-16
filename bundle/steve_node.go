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

package bundle

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// PatchNode patch a node described by req.Obj from steve server.
// Required fields: ClusterName, Name, Obj
func (b *Bundle) PatchNode(req *apistructs.SteveRequest) error {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName, name and type fields are required")
	}
	if !isObjInvalid(req.Obj) {
		return errors.New("obj in req is invalid")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return err
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1/node", req.Name)
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Patch(host).Path(path).Headers(headers).JSONBody(req.Obj).Do().RAW()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	return isSteveError(data)
}

// LabelNode labels a node.
// Required filed: ClusterName, Name
func (b *Bundle) LabelNode(req *apistructs.SteveRequest, labels map[string]string) error {
	if req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName and name fields are required")
	}

	if labels == nil || len(labels) == 0 {
		return errors.New("labels are required")
	}

	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
	}
	req.Obj = metadata
	return b.PatchNode(req)
}

// UnlabelNode unlabels a node.
// Required filed: ClusterName, Name
func (b *Bundle) UnlabelNode(req *apistructs.SteveRequest, labels []string) error {
	if req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName and name fields are required")
	}

	if len(labels) == 0 {
		return errors.New("labels are required")
	}

	toUnlabel := make(map[string]interface{})
	for _, label := range labels {
		toUnlabel[label] = nil
	}
	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": toUnlabel,
		},
	}
	req.Obj = metadata
	return b.PatchNode(req)
}

// CordonNode cordons a node.
// Required fields: ClusterName, Name
func (b *Bundle) CordonNode(req *apistructs.SteveRequest) error {
	if req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName and name fields are required")
	}

	spec := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": true,
		},
	}
	req.Obj = spec
	return b.PatchNode(req)
}

// UnCordonNode uncordons a node.
// Required fields: ClusterName, Name
func (b *Bundle) UnCordonNode(req *apistructs.SteveRequest) error {
	if req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName and name fields are required")
	}

	spec := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": nil,
		},
	}
	req.Obj = spec
	return b.PatchNode(req)
}
