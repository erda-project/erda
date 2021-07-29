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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// All methods here need bundle withCMP

// GetSteveResource gets k8s resource from steve server.
// Required fields: ClusterName, Name, Type.
func (b *Bundle) GetSteveResource(req *apistructs.SteveRequest) (*apistructs.SteveResource, error) {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return nil, errors.New("clusterName, name and type fields are required")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}

	// path format: /api/k8s/clusters/{clusterName}/v1/{type}/{namespace}/{name}
	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Get(host).Path(path).Headers(headers).Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if err = isSteveError(data); err != nil {
		return nil, err
	}

	var resource apistructs.SteveResource
	if err = json.Unmarshal(data, &resource); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &resource, nil
}

// ListSteveResource lists k8s resource from steve server.
// Required fields: ClusterName, Type.
func (b *Bundle) ListSteveResource(req *apistructs.SteveRequest) (*apistructs.SteveCollection, error) {
	if req.Type == "" || req.ClusterName == "" {
		return nil, errors.New("clusterName and type fields are required")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}

	// path format: /k8s/clusters/{clusterName}/v1/{type}/{namespace}?{label selectors}
	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace)
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Get(host).Path(path).Headers(headers).Params(req.URLQueryString()).Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if err = isSteveError(data); err != nil {
		return nil, err
	}

	var collection apistructs.SteveCollection
	if err = json.Unmarshal(data, &collection); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &collection, nil
}

// UpdateSteveResource update a k8s resource described by req.Obj from steve server.
// Required fields: ClusterName, Type, Name, Obj
func (b *Bundle) UpdateSteveResource(req *apistructs.SteveRequest) (*apistructs.SteveResource, error) {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return nil, errors.New("clusterName, name and type fields are required")
	}
	if !isObjInvalid(req.Obj) {
		return nil, errors.New("obj in req is invalid")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Put(host).Path(path).Headers(headers).JSONBody(req.Obj).Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if err = isSteveError(data); err != nil {
		return nil, err
	}

	var resource apistructs.SteveResource
	if err = json.Unmarshal(data, &resource); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &resource, nil
}

// CreateSteveResource creates a k8s resource described by req.Obj from steve server.
// Required fields: ClusterName, Type, Obj
func (b *Bundle) CreateSteveResource(req *apistructs.SteveRequest) (*apistructs.SteveResource, error) {
	if req.Type == "" || req.ClusterName == "" {
		return nil, errors.New("clusterName and type fields are required")
	}
	if !isObjInvalid(req.Obj) {
		return nil, errors.New("obj in req is invalid")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type))
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Post(host).Path(path).Headers(headers).JSONBody(req.Obj).Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if err = isSteveError(data); err != nil {
		return nil, err
	}

	var resource apistructs.SteveResource
	if err = json.Unmarshal(data, &resource); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &resource, nil
}

// DeleteSteveResource delete a k8s resource from steve server.
// Required fields: ClusterName, Type, Name
func (b *Bundle) DeleteSteveResource(req *apistructs.SteveRequest) error {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return errors.New("clusterName, name and type fields are required")
	}

	host, err := b.urls.CMP()
	if err != nil {
		return err
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)
	headers := http.Header{
		httputil.InternalHeader: []string{"bundle"},
		httputil.UserHeader:     []string{req.UserID},
		httputil.OrgHeader:      []string{req.OrgID},
	}
	hc := b.hc

	resp, err := hc.Delete(host).Path(path).Headers(headers).Do().RAW()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	return isSteveError(data)
}

func isObjInvalid(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr && !v.IsNil()
}

func isSteveError(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	typ, ok := obj["type"].(string)
	if !ok {
		return apierrors.ErrInvoke.InternalError(errors.New("type field is null"))
	}

	if typ != apistructs.SteveErrorType {
		return nil
	}

	var steveErr apistructs.SteveError
	if err = json.Unmarshal(data, &steveErr); err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	return toAPIError(steveErr.Status, apistructs.ErrorResponse{
		Code: steveErr.Code,
		Msg:  steveErr.Message,
	})
}
