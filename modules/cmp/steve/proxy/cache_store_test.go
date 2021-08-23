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

package proxy

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/attributes"
	v1 "github.com/rancher/wrangler/pkg/generated/controllers/rbac/v1"
	"github.com/rancher/wrangler/pkg/schemas"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/modules/cmp/cache"
)

type rbacInterface struct {
	v1.Interface
}

func (r *rbacInterface) Role() v1.RoleController {
	return &roleController{}
}

func (r *rbacInterface) RoleBinding() v1.RoleBindingController {
	return &roleBindingController{}
}

func (r *rbacInterface) ClusterRole() v1.ClusterRoleController {
	return &clusterRoleController{}
}

func (r *rbacInterface) ClusterRoleBinding() v1.ClusterRoleBindingController {
	return &clusterRoleBindingController{}
}

type roleController struct {
	v1.RoleController
}

func (r *roleController) Cache() v1.RoleCache {
	return &roleCache{}
}

func (r *roleController) OnChange(_ context.Context, _ string, _ v1.RoleHandler) {
}

type roleBindingController struct {
	v1.RoleBindingController
}

func (r *roleBindingController) Cache() v1.RoleBindingCache {
	return &roleBindingCache{}
}

type clusterRoleController struct {
	v1.ClusterRoleController
}

func (c *clusterRoleController) Cache() v1.ClusterRoleCache {
	return &clusterRoleCache{}
}

func (c *clusterRoleController) OnChange(_ context.Context, _ string, _ v1.ClusterRoleHandler) {
}

type clusterRoleBindingController struct {
	v1.ClusterRoleBindingController
}

func (c *clusterRoleBindingController) Cache() v1.ClusterRoleBindingCache {
	return &clusterRoleBindingCache{}
}

type roleCache struct {
	v1.RoleCache
}

func (r *roleCache) Get(_, name string) (*rbacv1.Role, error) {
	return testRoles[name], nil
}

type roleBindingCache struct {
	v1.RoleBindingCache
}

func (r *roleBindingCache) GetByIndex(_, key string) ([]*rbacv1.RoleBinding, error) {
	return testRoleBindings[key], nil
}

func (r *roleBindingCache) AddIndexer(_ string, _ v1.RoleBindingIndexer) {
}

type clusterRoleCache struct {
	v1.ClusterRoleCache
}

func (c *clusterRoleCache) Get(name string) (*rbacv1.ClusterRole, error) {
	return testClusterRoles[name], nil
}

type clusterRoleBindingCache struct {
	v1.ClusterRoleBindingCache
}

func (c *clusterRoleBindingCache) GetByIndex(_, key string) ([]*rbacv1.ClusterRoleBinding, error) {
	return testClusterRoleBindings[key], nil
}

func (c *clusterRoleBindingCache) AddIndexer(_ string, _ v1.ClusterRoleBindingIndexer) {
}

var (
	testRoles = map[string]*rbacv1.Role{
		// default namespace pods reader
		"viewer": {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Rules: []rbacv1.PolicyRule{
				{
					Verbs:     []string{"get"},
					APIGroups: []string{""},
					Resources: []string{"pods"},
				},
			},
		},
	}

	testRoleBindings = map[string][]*rbacv1.RoleBinding{
		// default namespace pods reader
		"viewer": {
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind: "Group",
						Name: "viewer",
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "Role",
					Name: "viewer",
				},
			},
		},
	}

	testClusterRoles = map[string]*rbacv1.ClusterRole{
		// test for admin
		"manager": {
			Rules: []rbacv1.PolicyRule{
				{
					Verbs:     []string{"*"},
					APIGroups: []string{"*"},
					Resources: []string{"*"},
				},
			},
		},
	}

	testClusterRoleBindings = map[string][]*rbacv1.ClusterRoleBinding{
		// test for admin
		"manager": {
			{
				Subjects: []rbacv1.Subject{
					{
						Kind: "Group",
						Name: "manager",
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "ClusterRole",
					Name: "manager",
				},
			},
		},
	}

	methods = []string{
		"get", "list", "update", "create", "update", "delete",
	}
)

func TestHasAccess(t *testing.T) {
	ctx := context.Background()
	cs := cacheStore{
		asl: accesscontrol.NewAccessStore(ctx, true, &rbacInterface{}),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://unit.test", nil)
	if err != nil {
		t.Error(err)
	}

	// test admin
	managerUser := &user.DefaultInfo{
		Name: "manager-01",
		UID:  "manager-01",
		Groups: []string{
			"manager",
		},
	}
	managerCtx := request.WithUser(ctx, managerUser)
	managerReq := &types.APIRequest{
		Request: req.WithContext(managerCtx),
	}
	for _, res := range []string{"roles", "clusterRoles", "roleBindings", "clusterRoleBindings", "secrets", "nodes"} {
		schema := &types.APISchema{
			Schema: &schemas.Schema{
				Attributes: map[string]interface{}{
					"group":    "",
					"resource": res,
				},
			},
		}
		for _, method := range methods {
			if !cs.hasAccess(managerReq, schema, method) {
				t.Errorf("test failed, user %s expected to have access for %s %s, actual not", method, managerUser.Name, res)
			}
		}
	}

	defaultPodsViewer := &user.DefaultInfo{
		Name: "defaultPodViewer-01",
		UID:  "defaultPodViewer-01",
		Groups: []string{
			"viewer",
		},
	}
	viewerCtx := request.WithUser(ctx, defaultPodsViewer)
	viewerReq := &types.APIRequest{
		Request: req.WithContext(viewerCtx),
	}
	for res, resAccess := range map[string]bool{
		"pods":         true,
		"deployments":  false,
		"statefulSets": false,
		"nodes":        false,
	} {
		schema := &types.APISchema{
			Schema: &schemas.Schema{
				Attributes: map[string]interface{}{
					"group":    "",
					"resource": res,
				},
			},
		}

		// all namespaces
		viewerReq.Namespace = ""
		if cs.hasAccess(viewerReq, schema, "get") {
			t.Errorf("test failed, user %s is not expected to have access for get %s in all namespaces, actual have", defaultPodsViewer.Name, res)
		}
		for i := 1; i < len(methods); i++ {
			if cs.hasAccess(viewerReq, schema, methods[i]) {
				t.Errorf("test failed, user %s is not expected to have access for %s %s, actual not", methods[i], defaultPodsViewer.Name, res)
			}
		}

		// default namespace
		viewerReq.Namespace = "default"
		if !cs.hasAccess(viewerReq, schema, "get") && resAccess {
			t.Errorf("test failed, user %s is expected to have access for get %s in default namespace, actual not", defaultPodsViewer.Name, res)
		} else if cs.hasAccess(viewerReq, schema, "get") && !resAccess {
			t.Errorf("test failed, user %s is not expected to have access for get %s in all namespaces, actual have", defaultPodsViewer.Name, res)
		}
	}
}

type store struct {
	types.Store
}

func (s *store) List(_ *types.APIRequest, _ *types.APISchema) (types.APIObjectList, error) {
	return types.APIObjectList{
		Revision: "-1",
		Continue: "false",
		Objects: []types.APIObject{
			{
				Type: "pod",
				ID:   "test",
			},
		},
	}, nil
}

func (s *store) Create(_ *types.APIRequest, _ *types.APISchema, _ types.APIObject) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (s *store) Update(_ *types.APIRequest, _ *types.APISchema, _ types.APIObject, _ string) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (s *store) Delete(_ *types.APIRequest, _ *types.APISchema, _ string) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func TestCacheStoreMethods(t *testing.T) {
	ctx := context.Background()
	cache, err := cache.New(256<<10, 256)
	if err != nil {
		t.Error(err)
	}
	cs := cacheStore{
		Store:   &store{},
		ctx:     ctx,
		asl:     accesscontrol.NewAccessStore(ctx, true, &rbacInterface{}),
		cache:   cache,
		cancels: sync.Map{},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://unit.test", nil)
	if err != nil {
		t.Error(err)
	}
	managerUser := &user.DefaultInfo{
		Name: "manager-02",
		UID:  "manager-02",
		Groups: []string{
			"manager",
		},
	}
	managerCtx := request.WithUser(ctx, managerUser)
	apiOp := &types.APIRequest{
		Request: req.WithContext(managerCtx),
	}

	schema := &types.APISchema{
		Schema: &schemas.Schema{
			Attributes: map[string]interface{}{
				"group":   "",
				"version": "v1",
				"kind":    "pod",
			},
		},
	}
	_, err = cs.List(apiOp, schema)
	if err != nil {
		t.Error(err)
	}

	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:       gvk.String(),
		namespace: "",
	}
	if res, _, err := cs.cache.Get(key.getKey()); res == nil || err != nil {
		t.Error("test failed, expected pods in cache, actual not")
	}

	_, err = cs.Create(apiOp, schema, types.APIObject{})
	if res, _, err := cs.cache.Get(key.getKey()); res != nil && err == nil {
		t.Error("test failed, expected no pods in cache, actual have")
	}

	_, err = cs.List(apiOp, schema)
	if err != nil {
		t.Error(err)
	}
	_, err = cs.Update(apiOp, schema, types.APIObject{}, "")
	if res, _, err := cs.cache.Get(key.getKey()); res != nil && err == nil {
		t.Error("test failed, expected no pods in cache, actual have")
	}

	_, err = cs.List(apiOp, schema)
	if err != nil {
		t.Error(err)
	}
	_, err = cs.Delete(apiOp, schema, "")
	if res, _, err := cs.cache.Get(key.getKey()); res != nil && err == nil {
		t.Error("test failed, expected no pods in cache, actual have")
	}
}
