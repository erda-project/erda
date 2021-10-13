// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	apiuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/modules/cmp/steve"
	"github.com/erda-project/erda/modules/cmp/steve/middleware"
	"github.com/erda-project/erda/pkg/strutil"
)

type SteveServer interface {
	GetSteveResource(context.Context, *apistructs.SteveRequest) (types.APIObject, error)
	ListSteveResource(context.Context, *apistructs.SteveRequest) ([]types.APIObject, error)
	UpdateSteveResource(context.Context, *apistructs.SteveRequest) (types.APIObject, error)
	CreateSteveResource(context.Context, *apistructs.SteveRequest) (types.APIObject, error)
	DeleteSteveResource(context.Context, *apistructs.SteveRequest) error
	PatchNode(context.Context, *apistructs.SteveRequest) error
	LabelNode(context.Context, *apistructs.SteveRequest, map[string]string) error
	UnlabelNode(context.Context, *apistructs.SteveRequest, []string) error
	CordonNode(context.Context, *apistructs.SteveRequest) error
	UnCordonNode(context.Context, *apistructs.SteveRequest) error
}

func (p *provider) GetSteveResource(ctx context.Context, req *apistructs.SteveRequest) (types.APIObject, error) {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return types.APIObject{}, errors.New("clusterName, name and type fields are required")
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return types.APIObject{}, err
	}
	if req.Type == apistructs.K8SNode {
		user = &apiuser.DefaultInfo{
			Name: "admin",
			UID:  "admin",
			Groups: []string{
				"system:masters",
				"system:authenticated",
			},
		}
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodGet, path, nil)
	if err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Name:           req.Name,
		Type:           string(req.Type),
		Method:         http.MethodGet,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return types.APIObject{}, err
	}

	rawRes, ok := resp.ResponseData.(*types.RawResource)
	if !ok {
		if resp.ResponseData == nil {
			return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
	}

	obj := rawRes.APIObject
	objData := obj.Data()
	if objData.String("type") == "error" {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New(objData.String("message")))
	}
	return obj, nil
}

func (p *provider) ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error) {
	if req.Type == "" || req.ClusterName == "" {
		return nil, apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and type fields are required"))
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace)

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return nil, err
	}
	if req.Type == apistructs.K8SNode {
		user = &apiuser.DefaultInfo{
			Name: "admin",
			UID:  "admin",
			Groups: []string{
				"system:masters",
				"system:authenticated",
			},
		}
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodGet, path, nil)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Name:           req.Name,
		Type:           string(req.Type),
		Method:         http.MethodGet,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}

	logrus.Infof("[DEBUG %s] start request steve aggregator at %s", req.Type, time.Now().Format(time.StampNano))
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	logrus.Infof("[DEBUG %s] end request steve aggregator at %s", req.Type, time.Now().Format(time.StampNano))

	collection, ok := resp.ResponseData.(*types.GenericCollection)
	if !ok {
		if resp.ResponseData == nil {
			return nil, apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		rawResource, ok := resp.ResponseData.(*types.RawResource)
		if !ok {
			return nil, apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
		}
		obj := rawResource.APIObject.Data()
		return nil, apierrors.ErrInvoke.InternalError(errors.New(obj.String("message")))
	}

	var objects []types.APIObject
	for _, obj := range collection.Data {
		objects = append(objects, obj.APIObject)
	}
	return objects, nil
}

func newReadCloser(obj interface{}) (io.ReadCloser, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(jsonData)), nil
}

func (p *provider) UpdateSteveResource(ctx context.Context, req *apistructs.SteveRequest) (types.APIObject, error) {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return types.APIObject{}, apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName, name and type fields are required"))
	}
	if !isObjInvalid(req.Obj) {
		return types.APIObject{}, apierrors.ErrInvoke.InvalidParameter(errors.New("obj in req is invalid"))
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return types.APIObject{}, err
	}

	body, err := newReadCloser(req.Obj)
	if err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InvalidParameter(errors.Errorf("failed to get body, %v", err))
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodPut, path, body)
	if err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Name:           req.Name,
		Type:           string(req.Type),
		Method:         http.MethodPut,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(err)
	}

	rawRes, ok := resp.ResponseData.(*types.RawResource)
	if !ok {
		if resp.ResponseData == nil {
			return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
	}

	obj := rawRes.APIObject
	objData := obj.Data()
	if objData.String("type") == "error" {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New(objData.String("message")))
	}

	auditCtx := map[string]interface{}{
		middleware.AuditClusterName:  req.ClusterName,
		middleware.AuditResourceType: req.Type,
		middleware.AuditNamespace:    req.Namespace,
		middleware.AuditResourceName: req.Name,
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditUpdateResource, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return obj, nil
}

func (p *provider) CreateSteveResource(ctx context.Context, req *apistructs.SteveRequest) (types.APIObject, error) {
	if req.Type == "" || req.ClusterName == "" {
		return types.APIObject{}, apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and type fields are required"))
	}
	if !isObjInvalid(req.Obj) {
		return types.APIObject{}, apierrors.ErrInvoke.InvalidParameter(errors.New("obj in req is invalid"))
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type))

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return types.APIObject{}, err
	}

	body, err := newReadCloser(req.Obj)
	if err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.Errorf("failed to get body, %v", err))
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodPost, path, body)
	if err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Type:           string(req.Type),
		Method:         http.MethodPost,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(err)
	}

	rawRes, ok := resp.ResponseData.(*types.RawResource)
	if !ok {
		if resp.ResponseData == nil {
			return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
	}

	obj := rawRes.APIObject
	objData := obj.Data()
	if objData.String("type") == "error" {
		return types.APIObject{}, apierrors.ErrInvoke.InternalError(errors.New(objData.String("message")))
	}

	reqObj, err := data.Convert(req.Obj)
	if err != nil {
		logrus.Errorf("failed to convert obj in request to data.Object, %v", err)
		return obj, nil
	}
	namespace := reqObj.String("metadata", "namespace")
	name := reqObj.String("metadata", "name")
	auditCtx := map[string]interface{}{
		middleware.AuditClusterName:  req.ClusterName,
		middleware.AuditResourceType: req.Type,
		middleware.AuditNamespace:    namespace,
		middleware.AuditResourceName: name,
	}

	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditCreateResource, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return obj, nil
}

func (p *provider) DeleteSteveResource(ctx context.Context, req *apistructs.SteveRequest) error {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName, name and type fields are required"))
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1", string(req.Type), req.Namespace, req.Name)

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return err
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodDelete, path, nil)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Name:           req.Name,
		Type:           string(req.Type),
		Method:         http.MethodDelete,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	rawRes, ok := resp.ResponseData.(*types.RawResource)
	if !ok {
		if resp.ResponseData == nil {
			return apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		return apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
	}

	obj := rawRes.APIObject
	objData := obj.Data()
	if objData.String("type") == "error" {
		return apierrors.ErrInvoke.InternalError(errors.New(objData.String("message")))
	}

	auditCtx := map[string]interface{}{
		middleware.AuditClusterName:  req.ClusterName,
		middleware.AuditResourceType: req.Type,
		middleware.AuditNamespace:    req.Namespace,
		middleware.AuditResourceName: req.Name,
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditDeleteResource, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return nil
}

func (p *provider) PatchNode(ctx context.Context, req *apistructs.SteveRequest) error {
	if req.Type == "" || req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName, name and type fields are required"))
	}
	if !isObjInvalid(req.Obj) {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("obj in req is invalid"))
	}

	path := strutil.JoinPath("/api/k8s/clusters", req.ClusterName, "v1/node", req.Name)

	user, err := p.Auth(req.UserID, req.OrgID, req.ClusterName)
	if err != nil {
		return err
	}

	body, err := newReadCloser(req.Obj)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(errors.Errorf("failed to get body, %v", err))
	}

	withUser := request.WithUser(ctx, user)
	r, err := http.NewRequestWithContext(withUser, http.MethodPatch, path, body)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	resp := &steve.Response{}
	apiOp := &types.APIRequest{
		Name:           req.Name,
		Type:           string(req.Type),
		Method:         http.MethodPatch,
		Namespace:      req.Namespace,
		ResponseWriter: resp,
		Request:        r,
		Response:       &steve.StatusCodeGetter{Response: resp},
	}
	if err := p.SteveAggregator.Serve(req.ClusterName, apiOp); err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	rawRes, ok := resp.ResponseData.(*types.RawResource)
	if !ok {
		if resp.ResponseData == nil {
			return apierrors.ErrInvoke.InternalError(errors.New("null response data"))
		}
		return apierrors.ErrInvoke.InternalError(errors.Errorf("unknown response data type: %s", reflect.TypeOf(resp.ResponseData).String()))
	}

	obj := rawRes.APIObject
	objData := obj.Data()
	if objData.String("type") == "error" {
		return apierrors.ErrInvoke.InternalError(errors.New(objData.String("message")))
	}
	return nil
}

func (p *provider) LabelNode(ctx context.Context, req *apistructs.SteveRequest, labels map[string]string) error {
	if req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and name fields are required"))
	}

	if labels == nil || len(labels) == 0 {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("labels are required"))
	}

	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
	}
	req.Obj = &metadata

	err := p.PatchNode(ctx, req)
	if err != nil {
		return err
	}

	var k, v string
	// there can only be one piece of k/v
	for k, v = range labels {
	}
	auditCtx := map[string]interface{}{
		middleware.AuditResourceName: req.Name,
		middleware.AuditTargetLabel:  fmt.Sprintf("%s=%s", k, v),
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditLabelNode, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return nil
}

func (p *provider) UnlabelNode(ctx context.Context, req *apistructs.SteveRequest, labels []string) error {
	if req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and name fields are required"))
	}

	if len(labels) == 0 {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("labels are required"))
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
	req.Obj = &metadata

	err := p.PatchNode(ctx, req)
	if err != nil {
		return err
	}

	auditCtx := map[string]interface{}{
		middleware.AuditResourceName: req.Name,
		middleware.AuditTargetLabel:  labels[0],
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditLabelNode, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return nil
}

func (p *provider) CordonNode(ctx context.Context, req *apistructs.SteveRequest) error {
	if req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and name fields are required"))
	}

	spec := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": true,
		},
	}
	req.Obj = &spec

	err := p.PatchNode(ctx, req)
	if err != nil {
		return err
	}

	auditCtx := map[string]interface{}{
		middleware.AuditResourceName: req.Name,
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditCordonNode, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return nil
}

func (p *provider) UnCordonNode(ctx context.Context, req *apistructs.SteveRequest) error {
	if req.ClusterName == "" || req.Name == "" {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("clusterName and name fields are required"))
	}

	spec := map[string]interface{}{
		"spec": map[string]interface{}{
			"unschedulable": false,
		},
	}
	req.Obj = &spec

	err := p.PatchNode(ctx, req)
	if err != nil {
		return err
	}

	auditCtx := map[string]interface{}{
		middleware.AuditResourceName: req.Name,
	}
	if err := p.Audit(req.UserID, req.OrgID, middleware.AuditUncordonNode, auditCtx); err != nil {
		logrus.Errorf("failed to audit when update steve resource, %v", err)
	}
	return nil
}

func (p *provider) Auth(userID, orgID, clusterName string) (apiuser.Info, error) {
	scopeID, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, apierrors.ErrInvoke.InvalidParameter(fmt.Sprintf("invalid org id %s, %v", orgID, err))
	}

	clusters, err := p.listClusterByType(scopeID, "k8s", "edas")
	if err != nil {
		return nil, err
	}

	found := false
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			found = true
			break
		}
	}
	if !found {
		return nil, apierrors.ErrInvoke.InvalidParameter(fmt.Sprintf("cluster %s not found in org %s", clusterName, orgID))
	}

	r := &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.OrgScope,
			ID:   orgID,
		},
	}
	rsp, err := p.SteveAggregator.Bdl.ScopeRoleAccess(userID, r)
	if err != nil {
		return nil, err
	}
	if !rsp.Access {
		return nil, apierrors.ErrInvoke.AccessDenied()
	}

	name := fmt.Sprintf("erda-user-%s", userID)
	user := &apiuser.DefaultInfo{
		Name: name,
		UID:  name,
	}
	for _, role := range rsp.Roles {
		group, ok := steve.RoleToGroup[role]
		if !ok {
			continue
		}
		user.Groups = append(user.Groups, group)
	}

	if len(user.Groups) == 0 {
		return nil, apierrors.ErrInvoke.AccessDenied()
	}

	return user, nil
}

func (p *provider) Audit(userID, orgID, templateName string, ctx map[string]interface{}) error {
	if ctx == nil || len(ctx) == 0 {
		return apierrors.ErrInvoke.InvalidParameter(errors.New("template context can not be empty"))
	}

	scopeID, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return apierrors.ErrInvoke.InvalidParameter(fmt.Errorf("invalid org id %s, %v", orgID, err))
	}

	now := strconv.FormatInt(time.Now().Unix(), 10)
	auditReq := apistructs.AuditCreateRequest{
		Audit: apistructs.Audit{
			UserID:       userID,
			ScopeType:    apistructs.OrgScope,
			ScopeID:      scopeID,
			OrgID:        scopeID,
			Context:      ctx,
			TemplateName: apistructs.TemplateName(templateName),
			Result:       "success",
			StartTime:    now,
			EndTime:      now,
		},
	}
	return p.SteveAggregator.Bdl.CreateAuditEvent(&auditReq)
}

func (p *provider) listClusterByType(orgID uint64, types ...string) ([]apistructs.ClusterInfo, error) {
	var result []apistructs.ClusterInfo
	for _, typ := range types {
		clusters, err := p.SteveAggregator.Bdl.ListClusters(typ, orgID)
		if err != nil {
			return nil, err
		}
		result = append(result, clusters...)
	}
	return result, nil
}

func isObjInvalid(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr && !v.IsNil()
}
